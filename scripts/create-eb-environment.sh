#!/bin/bash

# Variables from environment
APP_NAME="golang-restful-api"
ENV_NAME=${EB_ENV_NAME}
DOCKER_IMAGE=${DOCKER_IMAGE_NAME}
REGION=${AWS_REGION}
INSTANCE_PROFILE="ElasticBeanstalk-InstanceProfile"
SECURITY_GROUP_NAME=${SECURITY_GROUP_NAME}
VPC_ID=${VPC_ID}
SUBNET_ID=${SUBNET_ID}
VERSION_LABEL="v1"
SOLUTION_STACK_NAME="64bit Amazon Linux 2 v4.0.2 running Docker"
KEY_PAIR_NAME="ec2-keypair" 

#  Check required environment variables
if [ -z "$DOCKER_IMAGE" ]; then
    echo "Error: DOCKER_IMAGE is not set. Exiting."
    exit 1
fi

if [ -z "$ENV_NAME" ]; then
    echo "Error: EB_ENV_NAME is not set. Exiting."
    exit 1
fi

if [ -z "$VPC_ID" ] || [ -z "$SUBNET_ID" ]; then
    echo "Error: VPC_ID or SUBNET_ID is not set. Exiting."
    exit 1
fi

# Create or find a security group for Elastic Beanstalk
echo "Checking if security group $SECURITY_GROUP_NAME exists..."
security_group_id=$(aws ec2 describe-security-groups --filters Name=group-name,Values=$SECURITY_GROUP_NAME --query "SecurityGroups[0].GroupId" --output text --region $REGION)

if [ "$security_group_id" == "None" ]; then
    echo "Creating security group $SECURITY_GROUP_NAME..."
    security_group_id=$(aws ec2 create-security-group --group-name $SECURITY_GROUP_NAME --description "Security group for Elastic Beanstalk environment" --vpc-id $VPC_ID --region $REGION --query 'GroupId' --output text)

    echo "Adding inbound rules to security group $SECURITY_GROUP_NAME..."
    aws ec2 authorize-security-group-ingress --group-id $security_group_id --protocol tcp --port 80 --cidr 0.0.0.0/0 --region $REGION
    aws ec2 authorize-security-group-ingress --group-id $security_group_id --protocol tcp --port 5000 --cidr 0.0.0.0/0 --region $REGION
    aws ec2 authorize-security-group-ingress --group-id $security_group_id --protocol tcp --port 5000 --cidr ::/0 --region $REGION
    aws ec2 authorize-security-group-ingress --group-id $security_group_id --protocol tcp --port 22 --cidr 0.0.0.0/0 --region $REGION  # For SSH access
else
    echo "Security group $SECURITY_GROUP_NAME already exists with ID $security_group_id."
fi

# Create Elastic Beanstalk Application if not exists
if aws elasticbeanstalk describe-applications --application-names $APP_NAME --region $REGION | grep -q $APP_NAME; then
    echo "Elastic Beanstalk application $APP_NAME already exists."
else
    echo "Creating Elastic Beanstalk application $APP_NAME..."
    aws elasticbeanstalk create-application --application-name $APP_NAME --region $REGION
fi

# Create or update Elastic Beanstalk environment
env_exists=$(aws elasticbeanstalk describe-environments --application-name $APP_NAME --environment-names $ENV_NAME --query "Environments[0].Status" --output text --region $REGION)

if [ "$env_exists" != "None" ] && [ "$env_exists" != "Terminated" ]; then
    echo "Updating Elastic Beanstalk environment $ENV_NAME with Docker image from ECR..."
    aws elasticbeanstalk update-environment \
        --application-name $APP_NAME \
        --environment-name $ENV_NAME \
        --option-settings \
        Namespace=aws:elasticbeanstalk:container:docker,OptionName=ImageSourceUrl,Value="${DOCKER_IMAGE}" \
        Namespace=aws:elb:listener,OptionName=ListenerProtocol,Value=HTTP \
        Namespace=aws:elb:listener,OptionName=InstancePort,Value=5000 \
        Namespace=aws:elb:listener,OptionName=LoadBalancerPort,Value=80 \
        Namespace=aws:autoscaling:launchconfiguration,OptionName=IamInstanceProfile,Value=$INSTANCE_PROFILE \
        Namespace=aws:ec2:vpc,OptionName=VPCId,Value=$VPC_ID \
        Namespace=aws:ec2:vpc,OptionName=Subnets,Value=$SUBNET_ID \
        Namespace=aws:autoscaling:launchconfiguration,OptionName=SecurityGroups,Value=$security_group_id \
        Namespace=aws:autoscaling:launchconfiguration,OptionName=EC2KeyName,Value=$KEY_PAIR_NAME \
        --region $REGION
else
    echo "Creating Elastic Beanstalk environment with Docker image from ECR..."
    aws elasticbeanstalk create-environment \
        --application-name "$APP_NAME" \
        --environment-name "$ENV_NAME" \
        --solution-stack-name "$SOLUTION_STACK_NAME" \
        --option-settings \
        Namespace=aws:elasticbeanstalk:container:docker,OptionName=ImageSourceUrl,Value="${DOCKER_IMAGE}" \
        Namespace=aws:elb:listener,OptionName=ListenerProtocol,Value=HTTP \
        Namespace=aws:elb:listener,OptionName=InstancePort,Value=5000 \
        Namespace=aws:elb:listener,OptionName=LoadBalancerPort,Value=80 \
        Namespace=aws:autoscaling:launchconfiguration,OptionName=IamInstanceProfile,Value="$INSTANCE_PROFILE" \
        Namespace=aws:ec2:vpc,OptionName=VPCId,Value="$VPC_ID" \
        Namespace=aws:ec2:vpc,OptionName=Subnets,Value="$SUBNET_ID" \
        Namespace=aws:autoscaling:launchconfiguration,OptionName=SecurityGroups,Value="$security_group_id" \
        Namespace=aws:elasticbeanstalk:environment,OptionName=EnvironmentType,Value=LoadBalanced \
        Namespace=aws:elasticbeanstalk:cloudwatch:logs,OptionName=StreamLogs,Value=true \
        Namespace=aws:elasticbeanstalk:cloudwatch:logs,OptionName=DeleteOnTerminate,Value=true \
        Namespace=aws:elasticbeanstalk:cloudwatch:logs,OptionName=RetentionInDays,Value=14 \
        --region $REGION
fi

# Wait for environment to be ready with extended timeout (600 seconds, 15 retries)
echo "Waiting for the environment to be ready with extended timeout..."
MAX_ATTEMPTS=15  # Increase number of attempts
DELAY=40  # Increase the delay between retries to allow more time for the environment to update

for (( i=0; i<$MAX_ATTEMPTS; i++ ))
do
    health_status=$(aws elasticbeanstalk describe-environment-health --environment-name $ENV_NAME --attribute-names All --region $REGION --query "HealthStatus" --output text)
    
    if [ "$health_status" == "Ok" ]; then
        echo "Environment health status is Ok. Proceeding with the next step."
        break
    else
        echo "Attempt $((i+1))/$MAX_ATTEMPTS: Waiting for environment to be ready... (Health: $health_status)"
    fi
    
    if [ $i -eq $((MAX_ATTEMPTS-1)) ]; then
        echo "Max attempts exceeded. Deployment failed."
        exit 1
    fi
    
    sleep $DELAY
done

# Check environment health explicitly after waiting
if [ "$health_status" != "Ok" ]; then
    echo "Error: Environment health status is not Ok after waiting. Exiting."
    exit 1
fi

echo "Deployment to Elastic Beanstalk completed."

# Run security groups validation (enhancement for security scan)
echo "Validating Security Groups..."
aws ec2 describe-security-groups --group-ids $security_group_id --region $REGION || {
    echo "Error: Failed to validate security groups."
    exit 1
}

# Wait for environment to be fully ready before enabling CloudWatch logs
echo "Waiting for the environment to be in the Ready state..."

MAX_ATTEMPTS=10  # Maximum number of attempts to check readiness
DELAY=30  # Delay between each attempt in seconds

for (( i=0; i<$MAX_ATTEMPTS; i++ ))
do
    env_status=$(aws elasticbeanstalk describe-environments --environment-names $ENV_NAME --query "Environments[0].Status" --output text --region $REGION)

    if [ "$env_status" == "Ready" ]; then
        echo "Environment is in Ready state. Proceeding to enable CloudWatch logs."
        break
    else
        echo "Attempt $((i+1))/$MAX_ATTEMPTS: Environment is not Ready yet (Status: $env_status). Retrying in $DELAY seconds..."
    fi

    if [ $i -eq $((MAX_ATTEMPTS-1)) ]; then
        echo "Max attempts exceeded. Environment is not in Ready state. Exiting."
        exit 1
    fi

    sleep $DELAY
done

# Now that the environment is ready, enable CloudWatch logs
echo "Enabling CloudWatch Logs..."
aws elasticbeanstalk update-environment \
    --environment-name $ENV_NAME \
    --option-settings Namespace=aws:elasticbeanstalk:cloudwatch:logs,OptionName=StreamLogs,Value=true \
    --region $REGION

echo "CloudWatch logs have been successfully enabled."

echo "Deployment completed successfully."
