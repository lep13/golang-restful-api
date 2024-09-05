#!/bin/bash

# Variables from environment
APP_NAME="golang-restful-api"
ENV_NAME=${EB_ENV_NAME}
S3_BUCKET=${S3_BUCKET_NAME}
REGION=${AWS_REGION}
INSTANCE_PROFILE="ElasticBeanstalk-InstanceProfile"
SECURITY_GROUP_NAME=${SECURITY_GROUP_NAME}
VPC_ID=${VPC_ID}
SUBNET_ID=${SUBNET_ID}
VERSION_LABEL="v1"
SOLUTION_STACK_NAME="64bit Amazon Linux 2023 v4.1.3 running Go 1"
KEY_PAIR_NAME="my-ec2-keypair" 

# Check required environment variables
if [ -z "$S3_BUCKET" ]; then
    echo "Error: S3_BUCKET_NAME is not set. Exiting."
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

# Build the Go application before packaging
echo "Building the Go application..."
go build -o main . || {
    echo "Error: Failed to build the application. Ensure Go is installed and paths are set correctly. Exiting."
    exit 1
}

# Ensure the Procfile is present
echo "Creating Procfile..."
echo "web: ./main" > Procfile

# Update/Create .ebextensions directory for configuration settings
mkdir -p .ebextensions

cat <<EOL > .ebextensions/01_cloudwatch.config
option_settings:
  aws:elasticbeanstalk:cloudwatch:logs:
    StreamLogs: true
    DeleteOnTerminate: true
    RetentionInDays: 14
EOL

# Package application including the binary, Procfile, and .ebextensions directory
echo "Packaging the application..."
zip -r application.zip . main Procfile .ebextensions option-settings.json

# Create or find a security group for Elastic Beanstalk
echo "Checking if security group $SECURITY_GROUP_NAME exists..."
security_group_id=$(aws ec2 describe-security-groups --filters Name=group-name,Values=$SECURITY_GROUP_NAME --query "SecurityGroups[0].GroupId" --output text --region $REGION)

if [ "$security_group_id" == "None" ]; then
    echo "Creating security group $SECURITY_GROUP_NAME..."
    security_group_id=$(aws ec2 create-security-group --group-name $SECURITY_GROUP_NAME --description "Security group for Elastic Beanstalk environment" --vpc-id $VPC_ID --region $REGION --query 'GroupId' --output text)
    
    echo "Adding inbound rules to security group $SECURITY_GROUP_NAME..."
    aws ec2 authorize-security-group-ingress --group-id $security_group_id --protocol tcp --port 5000 --cidr 0.0.0.0/0 --region $REGION
    aws ec2 authorize-security-group-ingress --group-id $security_group_id --protocol tcp --port 5000 --cidr ::/0 --region $REGION
    aws ec2 authorize-security-group-ingress --group-id $security_group_id --protocol tcp --port 443 --cidr 0.0.0.0/0 --region $REGION
    aws ec2 authorize-security-group-ingress --group-id $security_group_id --protocol tcp --port 22 --cidr 0.0.0.0/0 --region $REGION  # For SSH access
else
    echo "Security group $SECURITY_GROUP_NAME already exists with ID $security_group_id."
fi

# Create S3 bucket if not exists
if aws s3api head-bucket --bucket "$S3_BUCKET" --region $REGION 2>/dev/null; then
    echo "S3 bucket $S3_BUCKET already exists."
else
    echo "Creating S3 bucket $S3_BUCKET..."
    aws s3 mb s3://$S3_BUCKET --region $REGION
fi

# Upload package to S3
echo "Uploading application.zip to S3..."
aws s3 cp application.zip s3://$S3_BUCKET/application.zip --region $REGION

# Create Elastic Beanstalk Application if not exists
if aws elasticbeanstalk describe-applications --application-names $APP_NAME --region $REGION | grep -q $APP_NAME; then
    echo "Elastic Beanstalk application $APP_NAME already exists."
else
    echo "Creating Elastic Beanstalk application $APP_NAME..."
    aws elasticbeanstalk create-application --application-name $APP_NAME --region $REGION
fi

# Create or update application version
echo "Creating new application version..."
aws elasticbeanstalk create-application-version --application-name $APP_NAME --version-label $VERSION_LABEL --source-bundle S3Bucket="$S3_BUCKET",S3Key="application.zip" --region $REGION || {
    echo "Application version $VERSION_LABEL already exists. Skipping creation."
}

# Check if environment exists and update or create accordingly
# Check if environment exists and update or create accordingly
env_exists=$(aws elasticbeanstalk describe-environments --application-name $APP_NAME --environment-names $ENV_NAME --query "Environments[0].Status" --output text --region $REGION)

if [ "$env_exists" != "None" ] && [ "$env_exists" != "Terminated" ]; then
    echo "Updating Elastic Beanstalk environment $ENV_NAME..."
    aws elasticbeanstalk update-environment \
        --application-name $APP_NAME \
        --environment-name $ENV_NAME \
        --version-label v1 \
        --option-settings Namespace=aws:autoscaling:launchconfiguration,OptionName=IamInstanceProfile,Value=$INSTANCE_PROFILE \
        Namespace=aws:ec2:vpc,OptionName=VPCId,Value=$VPC_ID \
        Namespace=aws:ec2:vpc,OptionName=Subnets,Value=$SUBNET_ID \
        Namespace=aws:autoscaling:launchconfiguration,OptionName=SecurityGroups,Value=$security_group_id \
        Namespace=aws:autoscaling:launchconfiguration,OptionName=EC2KeyName,Value=$KEY_PAIR_NAME \
        Namespace=aws:elasticbeanstalk:environment,OptionName=LoadBalancerHTTPPort,Value=80 \
        Namespace=aws:elasticbeanstalk:environment,OptionName=InstancePort,Value=5000 \
        --region $REGION
else
    echo "Creating Elastic Beanstalk environment $ENV_NAME..."
    aws elasticbeanstalk create-environment \
        --application-name "$APP_NAME" \
        --environment-name "$ENV_NAME" \
        --version-label "v1" \
        --solution-stack-name "$SOLUTION_STACK_NAME" \
        --option-settings Namespace=aws:autoscaling:launchconfiguration,OptionName=IamInstanceProfile,Value="$INSTANCE_PROFILE" \
        Namespace=aws:ec2:vpc,OptionName=VPCId,Value="$VPC_ID" \
        Namespace=aws:ec2:vpc,OptionName=Subnets,Value="$SUBNET_ID" \
        Namespace=aws:autoscaling:launchconfiguration,OptionName=SecurityGroups,Value="$security_group_id" \
        Namespace=aws:elasticbeanstalk:environment,OptionName=EnvironmentType,Value=LoadBalanced \
        Namespace=aws:elasticbeanstalk:cloudwatch:logs,OptionName=StreamLogs,Value=true \
        Namespace=aws:elasticbeanstalk:cloudwatch:logs,OptionName=DeleteOnTerminate,Value=true \
        Namespace=aws:elasticbeanstalk:cloudwatch:logs,OptionName=RetentionInDays,Value=14 \
        Namespace=aws:autoscaling:launchconfiguration,OptionName=EC2KeyName,Value=$KEY_PAIR_NAME \
        Namespace=aws:elasticbeanstalk:environment,OptionName=LoadBalancerHTTPPort,Value=80 \
        Namespace=aws:elasticbeanstalk:environment,OptionName=InstancePort,Value=5000 \
        --region $REGION || {
            echo "Error: Failed to create Elastic Beanstalk environment. Exiting."
            exit 1
        }
fi

# Wait for environment to be ready
echo "Waiting for the environment to be ready..."
aws elasticbeanstalk wait environment-updated --application-name $APP_NAME --environment-names $ENV_NAME --region $REGION || {
    echo "Error: Environment update did not complete successfully. Exiting."
    exit 1
}

# Check environment health and perform health check
echo "Checking environment health..."
health_status=$(aws elasticbeanstalk describe-environment-health --environment-name $ENV_NAME --attribute-names All --region $REGION --query "HealthStatus" --output text)
if [ "$health_status" != "Ok" ]; then
    echo "Environment health status: $health_status. Please investigate further."
else
    echo "Environment health status is Ok."
fi

# Enable CloudWatch monitoring and logs if not enabled
echo "Enabling CloudWatch monitoring and logs..."
aws elasticbeanstalk update-environment \
    --environment-name $ENV_NAME \
    --option-settings Namespace=aws:elasticbeanstalk:cloudwatch:logs,OptionName=StreamLogs,Value=true \
    --region $REGION

echo "Deployment to Elastic Beanstalk completed successfully."
