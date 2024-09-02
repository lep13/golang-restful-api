#!/bin/bash

# Variables from environment
APP_NAME="golang-restful-api"
ENV_NAME=${EB_ENV_NAME}
S3_BUCKET=${S3_BUCKET_NAME}
REGION=${AWS_REGION}
VERSION_LABEL="v1"

# Check if environment variables are set
if [ -z "$S3_BUCKET" ]; then
    echo "Error: S3_BUCKET_NAME is not set. Exiting."
    exit 1
fi

if [ -z "$ENV_NAME" ]; then
    echo "Error: EB_ENV_NAME is not set. Exiting."
    exit 1
fi

if [ -z "$REGION" ]; then
    echo "Error: AWS_REGION is not set. Exiting."
    exit 1
fi

# Create S3 bucket if not exists
if aws s3api head-bucket --bucket "$S3_BUCKET" 2>/dev/null; then
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
aws s3 cp application.zip s3://$S3_BUCKET/application.zip
echo "Checking contents of S3 bucket..."
aws s3 ls s3://$S3_BUCKET/

# Create Elastic Beanstalk Application if not exists
if aws elasticbeanstalk describe-applications --application-names $APP_NAME | grep -q $APP_NAME; then
    echo "Elastic Beanstalk application $APP_NAME already exists."
else
    echo "Creating Elastic Beanstalk application $APP_NAME..."
    aws elasticbeanstalk create-application --application-name $APP_NAME --region $REGION
fi

# Create new application version
echo "Creating new application version..."
aws elasticbeanstalk create-application-version \
    --application-name $APP_NAME \
    --version-label $VERSION_LABEL \
    --source-bundle S3Bucket="$S3_BUCKET",S3Key="application.zip" \
    --region $REGION

# Check if application version is created
echo "Checking application versions..."
aws elasticbeanstalk describe-application-versions --application-name $APP_NAME --region $REGION

# Create or Update Elastic Beanstalk Environment
existing_env=$(aws elasticbeanstalk describe-environments --application-name $APP_NAME --environment-names $ENV_NAME --query 'Environments[0].EnvironmentName' --output text)

if [ "$existing_env" == "$ENV_NAME" ]; then
    echo "Updating existing Elastic Beanstalk environment $ENV_NAME..."
    aws elasticbeanstalk update-environment --environment-name $ENV_NAME --version-label $VERSION_LABEL --region $REGION
else
    echo "Creating new Elastic Beanstalk environment $ENV_NAME..."
    aws elasticbeanstalk create-environment \
        --application-name $APP_NAME \
        --environment-name $ENV_NAME \
        --solution-stack-name "64bit Amazon Linux 2 v3.3.8 running Go 1.x" \
        --version-label $VERSION_LABEL \
        --option-settings Namespace=aws:elasticbeanstalk:environment,OptionName=EnvironmentType,Value=SingleInstance \
        --option-settings Namespace=aws:elasticbeanstalk:application:environment,OptionName=MONGO_URI,Value=${MONGO_URI} \
        --option-settings Namespace=aws:elasticbeanstalk:application:environment,OptionName=DB_NAME,Value=${DB_NAME} \
        --region $REGION
fi

echo "Deployment to Elastic Beanstalk completed."
