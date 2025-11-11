package web

import (
	"encoding/base64"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/alb"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/autoscaling"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/cloudwatch"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/route53"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// WebEnvironmentArgs defines the input arguments for the WebEnvironment component
type WebEnvironmentArgs struct {
	VpcId            pulumi.StringInput      `pulumi:"vpcId"`
	VpcCidr          pulumi.StringInput      `pulumi:"vpcCidr"`
	ImageId          pulumi.StringInput      `pulumi:"imageId"`
	InstanceCount    pulumi.IntPtrInput      `pulumi:"instanceCount,optional"`
	InstanceType     pulumi.StringInput      `pulumi:"instanceType"`
	PublicSubnetIds  pulumi.StringArrayInput `pulumi:"publicSubnetIds"`
	PrivateSubnetIds pulumi.StringArrayInput `pulumi:"privateSubnetIds,optional"`
	CertificateArn   pulumi.StringInput      `pulumi:"certificateArn"`
	ZoneId           pulumi.StringPtrInput   `pulumi:"zoneId,optional"`
	Subdomain        pulumi.StringPtrInput   `pulumi:"subdomain,optional"`
}

// WebEnvironment is a component that creates a complete web hosting environment
type WebEnvironment struct {
	pulumi.ResourceState

	LoadBalancerUrl pulumi.StringOutput `pulumi:"loadBalancerUrl"`
}

// NewWebEnvironment creates a new WebEnvironment component
func NewWebEnvironment(ctx *pulumi.Context, name string, args WebEnvironmentArgs, opts ...pulumi.ResourceOption) (*WebEnvironment, error) {
	comp := &WebEnvironment{}

	// Register the component resource
	err := ctx.RegisterComponentResource("web-components:index:WebEnvironment", name, comp, opts...)
	if err != nil {
		return nil, err
	}

	// Create ALB security group
	albSg, err := ec2.NewSecurityGroup(ctx, name+"-alb-sg", &ec2.SecurityGroupArgs{
		VpcId: args.VpcId,
		Ingress: ec2.SecurityGroupIngressArray{
			&ec2.SecurityGroupIngressArgs{
				Protocol:   pulumi.String("tcp"),
				FromPort:   pulumi.Int(443),
				ToPort:     pulumi.Int(443),
				CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			},
			&ec2.SecurityGroupIngressArgs{
				Protocol:   pulumi.String("tcp"),
				FromPort:   pulumi.Int(80),
				ToPort:     pulumi.Int(80),
				CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			},
		},
		Egress: ec2.SecurityGroupEgressArray{
			&ec2.SecurityGroupEgressArgs{
				Protocol:   pulumi.String("-1"),
				FromPort:   pulumi.Int(0),
				ToPort:     pulumi.Int(0),
				CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			},
		},
	}, pulumi.Parent(comp), pulumi.RetainOnDelete(true), pulumi.Protect(true))
	if err != nil {
		return nil, err
	}

	// Create instance security group
	instanceSg, err := ec2.NewSecurityGroup(ctx, name+"-instance-sg", &ec2.SecurityGroupArgs{
		VpcId: args.VpcId,
		Ingress: ec2.SecurityGroupIngressArray{
			&ec2.SecurityGroupIngressArgs{
				Protocol: pulumi.String("tcp"),
				FromPort: pulumi.Int(22),
				ToPort:   pulumi.Int(22),
				CidrBlocks: args.VpcCidr.ToStringOutput().ApplyT(func(cidr string) []string {
					return []string{cidr}
				}).(pulumi.StringArrayOutput),
			},
			&ec2.SecurityGroupIngressArgs{
				Protocol:       pulumi.String("tcp"),
				FromPort:       pulumi.Int(80),
				ToPort:         pulumi.Int(80),
				SecurityGroups: pulumi.StringArray{albSg.ID()},
			},
		},
		Egress: ec2.SecurityGroupEgressArray{
			&ec2.SecurityGroupEgressArgs{
				Protocol:   pulumi.String("-1"),
				FromPort:   pulumi.Int(0),
				ToPort:     pulumi.Int(0),
				CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			},
		},
	}, pulumi.Parent(comp))
	if err != nil {
		return nil, err
	}

	// Create SSH key pair
	privateKey, err := tls.NewPrivateKey(ctx, name, &tls.PrivateKeyArgs{
		Algorithm: pulumi.String("RSA"),
	}, pulumi.Parent(comp))
	if err != nil {
		return nil, err
	}

	sshKey, err := ec2.NewKeyPair(ctx, name, &ec2.KeyPairArgs{
		PublicKey: privateKey.PublicKeyOpenssh,
	}, pulumi.Parent(privateKey))
	if err != nil {
		return nil, err
	}

	// Create launch template
	userData := `#!/bin/bash
sudo yum update -y
sudo amazon-linux-extras install nginx1 -y
sudo systemctl enable nginx
sudo systemctl start nginx`
	encodedUserData := base64.StdEncoding.EncodeToString([]byte(userData))

	launchTemplate, err := ec2.NewLaunchTemplate(ctx, "launch-config", &ec2.LaunchTemplateArgs{
		NamePrefix:   pulumi.String(name + "-web"),
		InstanceType: args.InstanceType,
		ImageId:      args.ImageId,
		KeyName:      sshKey.KeyName,
		VpcSecurityGroupIds: pulumi.StringArray{
			instanceSg.ID(),
		},
		Tags: pulumi.StringMap{
			"Role": pulumi.String("Webserver"),
		},
		UserData: pulumi.String(encodedUserData),
	}, pulumi.Parent(comp))
	if err != nil {
		return nil, err
	}

	// Create auto scaling group
	asg, err := autoscaling.NewGroup(ctx, name+"-asg", &autoscaling.GroupArgs{
		VpcZoneIdentifiers: args.PublicSubnetIds,
		DesiredCapacity:    pulumi.Int(1),
		MaxSize:            pulumi.Int(1),
		MinSize:            pulumi.Int(1),
		LaunchTemplate: &autoscaling.GroupLaunchTemplateArgs{
			Id:      launchTemplate.ID(),
			Version: pulumi.String("$Latest"),
		},
	}, pulumi.Parent(launchTemplate))
	if err != nil {
		return nil, err
	}

	// Create application load balancer
	loadBalancer, err := alb.NewLoadBalancer(ctx, name+"-alb", &alb.LoadBalancerArgs{
		Internal:       pulumi.Bool(false),
		SecurityGroups: pulumi.StringArray{albSg.ID()},
		Subnets:        args.PublicSubnetIds,
	}, pulumi.Parent(comp))
	if err != nil {
		return nil, err
	}

	// Create scaling policies
	scaleOutPolicy, err := autoscaling.NewPolicy(ctx, "scaleOutPolicy", &autoscaling.PolicyArgs{
		AdjustmentType:       pulumi.String("ChangeInCapacity"),
		ScalingAdjustment:    pulumi.Int(1),
		AutoscalingGroupName: asg.Name,
	}, pulumi.Parent(asg))
	if err != nil {
		return nil, err
	}

	scaleInPolicy, err := autoscaling.NewPolicy(ctx, "scaleInPolicy", &autoscaling.PolicyArgs{
		AdjustmentType:       pulumi.String("ChangeInCapacity"),
		ScalingAdjustment:    pulumi.Int(-1),
		Cooldown:             pulumi.Int(180),
		AutoscalingGroupName: asg.Name,
	}, pulumi.Parent(asg))
	if err != nil {
		return nil, err
	}

	// Create CloudWatch alarms
	_, err = cloudwatch.NewMetricAlarm(ctx, "albHighRequestAlarm", &cloudwatch.MetricAlarmArgs{
		Name:       pulumi.String("alb-high-requests-alarm"),
		Namespace:  pulumi.String("AWS/ApplicationELB"),
		MetricName: pulumi.String("RequestCount"),
		Dimensions: pulumi.StringMap{
			"LoadBalancer": pulumi.String(name + "-alb"),
		},
		Period:             pulumi.Int(180),
		EvaluationPeriods:  pulumi.Int(2),
		Threshold:          pulumi.Float64(1000.0),
		ComparisonOperator: pulumi.String("GreaterThanOrEqualToThreshold"),
		Statistic:          pulumi.String("Sum"),
		AlarmActions:       pulumi.Array{scaleOutPolicy.Arn}.ToArrayOutput(),
	}, pulumi.Parent(scaleOutPolicy))
	if err != nil {
		return nil, err
	}

	_, err = cloudwatch.NewMetricAlarm(ctx, "albLowRequestAlarm", &cloudwatch.MetricAlarmArgs{
		Name:       pulumi.String("alb-low-requests-alarm"),
		Namespace:  pulumi.String("AWS/ApplicationELB"),
		MetricName: pulumi.String("RequestCount"),
		Dimensions: pulumi.StringMap{
			"LoadBalancer": pulumi.String(name + "-alb"),
		},
		Period:             pulumi.Int(180),
		EvaluationPeriods:  pulumi.Int(2),
		Threshold:          pulumi.Float64(500.0),
		ComparisonOperator: pulumi.String("LessThanOrEqualToThreshold"),
		Statistic:          pulumi.String("Sum"),
		AlarmActions:       pulumi.Array{scaleInPolicy.Arn}.ToArrayOutput(),
	}, pulumi.Parent(scaleInPolicy))
	if err != nil {
		return nil, err
	}

	// Create target group
	tg, err := alb.NewTargetGroup(ctx, name+"-tg", &alb.TargetGroupArgs{
		TargetType: pulumi.String("instance"),
		Port:       pulumi.Int(80),
		Protocol:   pulumi.String("HTTP"),
		VpcId:      args.VpcId,
	}, pulumi.Parent(loadBalancer))
	if err != nil {
		return nil, err
	}

	// Create HTTPS listener
	_, err = alb.NewListener(ctx, name+"-frontend-https", &alb.ListenerArgs{
		LoadBalancerArn: loadBalancer.Arn,
		Port:            pulumi.Int(443),
		Protocol:        pulumi.String("HTTPS"),
		CertificateArn:  args.CertificateArn,
		DefaultActions: alb.ListenerDefaultActionArray{
			&alb.ListenerDefaultActionArgs{
				Type:           pulumi.String("forward"),
				TargetGroupArn: tg.Arn,
			},
		},
	}, pulumi.Parent(loadBalancer))
	if err != nil {
		return nil, err
	}

	// Create HTTP listener with redirect
	_, err = alb.NewListener(ctx, name+"-frontend-redir", &alb.ListenerArgs{
		LoadBalancerArn: loadBalancer.Arn,
		Port:            pulumi.Int(80),
		Protocol:        pulumi.String("HTTP"),
		DefaultActions: alb.ListenerDefaultActionArray{
			&alb.ListenerDefaultActionArgs{
				Type: pulumi.String("redirect"),
				Redirect: &alb.ListenerDefaultActionRedirectArgs{
					Protocol:   pulumi.String("HTTPS"),
					Port:       pulumi.String("443"),
					StatusCode: pulumi.String("HTTP_301"),
				},
			},
		},
	}, pulumi.Parent(loadBalancer))
	if err != nil {
		return nil, err
	}

	// Attach ASG to target group
	_, err = autoscaling.NewAttachment(ctx, name+"-alb-att", &autoscaling.AttachmentArgs{
		AutoscalingGroupName: asg.Name,
		LbTargetGroupArn:     tg.Arn,
	}, pulumi.Parent(asg))
	if err != nil {
		return nil, err
	}

	// Create Route53 alias record if zone ID and subdomain are provided
	if args.ZoneId != nil && args.Subdomain != nil {
		_, err = route53.NewRecord(ctx, "alias", &route53.RecordArgs{
			ZoneId: args.ZoneId.ToStringPtrOutput().Elem(),
			Name:   args.Subdomain.ToStringPtrOutput().Elem(),
			Type:   pulumi.String("A"),
			Aliases: route53.RecordAliasArray{
				&route53.RecordAliasArgs{
					Name:                 loadBalancer.DnsName,
					ZoneId:               loadBalancer.ZoneId,
					EvaluateTargetHealth: pulumi.Bool(true),
				},
			},
		}, pulumi.Parent(comp))
		if err != nil {
			return nil, err
		}
	}

	// Set the load balancer URL output
	comp.LoadBalancerUrl = loadBalancer.DnsName.ApplyT(func(dns string) string {
		return "https://" + dns
	}).(pulumi.StringOutput)

	// Register outputs
	if err := ctx.RegisterResourceOutputs(comp, pulumi.Map{
		"loadBalancerUrl": comp.LoadBalancerUrl,
	}); err != nil {
		return nil, err
	}

	return comp, nil
}

// Annotate provides descriptions for the component and its properties
func (w *WebEnvironmentArgs) Annotate(a infer.Annotator) {
	a.Describe(&w.VpcId, "The VPC ID where resources will be created")
	a.Describe(&w.VpcCidr, "The CIDR block of the VPC for security group rules")
	a.Describe(&w.ImageId, "The AMI ID to use for EC2 instances")
	a.Describe(&w.InstanceCount, "The number of instances to run (optional)")
	a.Describe(&w.InstanceType, "The EC2 instance type")
	a.Describe(&w.PublicSubnetIds, "List of public subnet IDs for the load balancer and instances")
	a.Describe(&w.PrivateSubnetIds, "List of private subnet IDs (optional)")
	a.Describe(&w.CertificateArn, "The ARN of the ACM certificate for HTTPS")
	a.Describe(&w.ZoneId, "The Route53 hosted zone ID for DNS alias (optional)")
	a.Describe(&w.Subdomain, "The subdomain for the DNS alias record (optional)")
}

// Annotate provides descriptions for the component and its properties
func (w *WebEnvironment) Annotate(a infer.Annotator) {
	a.Describe(&w, "Creates a complete web hosting environment with auto-scaling, load balancing, and optional DNS configuration")
	a.Describe(&w.LoadBalancerUrl, "The HTTPS URL of the Application Load Balancer")
}
