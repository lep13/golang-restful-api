#!/bin/bash

# Variables from environment
APP_NAME="golang-restful-api"
ENV_NAME=${EB_ENV_NAME}
S3_BUCKET=${S3_BUCKET_NAME}
REGION=${AWS_REGION}
INSTANCE_PROFILE=${IAM_INSTANCE_PROFILE}  # Use the IAM instance profile from secrets
SECURITY_GROUP_NAME=${SECURITY_GROUP_NAME}  # Use the security group name from secrets
VPC_ID=${VPC_ID}
SUBNET_ID=${SUBNET_ID}
VERSION_LABEL="v1"  # Define the version label
SOLUTION_STACK_NAME="64bit Amazon Linux 2023 v4.1.3 running Go 1"

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

# Create or find a security group for Elastic Beanstalk
echo "Checking if security group $SECURITY_GROUP_NAME exists..."
security_group_id=$(aws ec2 describe-security-groups --filters Name=group-name,Values=$SECURITY_GROUP_NAME --query "SecurityGroups[0].GroupId" --output text --region $REGION)

if [ "$security_group_id" == "None" ]; then
    echo "Creating security group $SECURITY_GROUP_NAME..."
    security_group_id=$(aws ec2 create-security-group --group-name $SECURITY_GROUP_NAME --description "Security group for Elastic Beanstalk environment" --vpc-id $VPC_ID --region $REGION --query 'GroupId' --output text)
    
    echo "Adding inbound rules to security group $SECURITY_GROUP_NAME..."
    aws ec2 authorize-security-group-ingress --group-id $security_group_id --protocol tcp --port 80 --cidr 0.0.0.0/0 --region $REGION
    aws ec2 authorize-security-group-ingress --group-id $security_group_id --protocol tcp --port 443 --cidr 0.0.0.0/0 --region $REGION
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

# Package application
echo "Packaging the application..."
zip -r application.zip .

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

# Create new application version
echo "Creating new application version..."
aws elasticbeanstalk create-application-version --application-name $APP_NAME --version-label $VERSION_LABEL --source-bundle S3Bucket="$S3_BUCKET",S3Key="application.zip" --region $REGION

# Create or update Elastic Beanstalk Environment with IAM Instance Profile and Security Group
echo "Checking if environment $ENV_NAME exists..."
# Check if environment exists
env_exists=$(aws elasticbeanstalk describe-environments --application-name $APP_NAME --environment-names $ENV_NAME --region $REGION --query "Environments[0].EnvironmentName" --output text)

if [ "$env_exists" == "None" ]; then
    echo "Creating Elastic Beanstalk environment $ENV_NAME..."
    aws elasticbeanstalk create-environment \
        --application-name $APP_NAME \
        --environment-name $ENV_NAME \
        --version-label $(date +%Y%m%d%H%M%S) \  # Use a timestamp to ensure unique version
        --solution-stack-name "64bit Amazon Linux 2023 v4.1.3 running Go 1" \
        --option-settings Namespace=aws:autoscaling:launchconfiguration,OptionName=IamInstanceProfile,Value=$INSTANCE_PROFILE \
        --option-settings Namespace=aws:ec2:vpc,OptionName=VPCId,Value=$VPC_ID \
        --option-settings Namespace=aws:ec2:vpc,OptionName=Subnets,Value=$SUBNET_ID \
        --option-settings Namespace=aws:autoscaling:launchconfiguration,OptionName=SecurityGroups,Value=$security_group_id \
        --region $REGION
else
    echo "Updating Elastic Beanstalk environment $ENV_NAME..."
    aws elasticbeanstalk update-environment \
        --environment-name $ENV_NAME \
        --option-settings Namespace=aws:autoscaling:launchconfiguration,OptionName=IamInstanceProfile,Value=$INSTANCE_PROFILE \
        --option-settings Namespace=aws:ec2:vpc,OptionName=VPCId,Value=$VPC_ID \
        --option-settings Namespace=aws:ec2:vpc,OptionName=Subnets,Value=$SUBNET_ID \
        --option-settings Namespace=aws:autoscaling:launchconfiguration,OptionName=SecurityGroups,Value=$security_group_id \
        --region $REGION
fi
    # Wait for the environment to become ready
    aws elasticbeanstalk wait environment-exists --environment-names $ENV_NAME --region $REGION
    echo "Environment $ENV_NAME is now in launching state."

    # Loop to check environment status
    while [ "$(aws elasticbeanstalk describe-environments --environment-names $ENV_NAME --query "Environments[0].Health" --output text --region $REGION)" != "Green" ]
    do
        echo "Waiting for environment $ENV_NAME to be Green..."
        sleep 30
    done
    echo "Environment $ENV_NAME is Green. Proceeding with deployment."
fi

echo "Deployment to Elastic Beanstalk completed."
