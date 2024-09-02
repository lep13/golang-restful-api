#!/bin/bash

# Ensure required environment variables are set
if [[ -z "$S3_BUCKET_NAME" || -z "$EB_ENV_NAME" ]]; then
    echo "Error: S3_BUCKET_NAME and EB_ENV_NAME must be set. Exiting."
    exit 1
fi

# Variables from environment
APP_NAME="golang-restful-api"
ENV_NAME="$EB_ENV_NAME"
S3_BUCKET="$S3_BUCKET_NAME"
REGION="$AWS_REGION"

# Create S3 bucket if it does not exist
if aws s3api head-bucket --bucket "$S3_BUCKET" 2>/dev/null; then
    echo "S3 bucket $S3_BUCKET already exists."
else
    echo "Creating S3 bucket $S3_BUCKET..."
    aws s3 mb s3://$S3_BUCKET --region $REGION || { echo "Failed to create S3 bucket. Exiting."; exit 1; }
fi

# Package application
echo "Packaging the application..."
zip -r application.zip . -x "*.git*"

# Upload package to S3
echo "Uploading application.zip to S3..."
aws s3 cp application.zip s3://$S3_BUCKET/application.zip || { echo "Failed to upload to S3. Exiting."; exit 1; }

# Create Elastic Beanstalk Application if not exists
if aws elasticbeanstalk describe-applications --application-names $APP_NAME | grep -q $APP_NAME; then
    echo "Elastic Beanstalk application $APP_NAME already exists."
else
    echo "Creating Elastic Beanstalk application $APP_NAME..."
    aws elasticbeanstalk create-application --application-name $APP_NAME --region $REGION || { echo "Failed to create Elastic Beanstalk application. Exiting."; exit 1; }
fi

# Create new application version
echo "Creating new application version..."
aws elasticbeanstalk create-application-version --application-name $APP_NAME --version-label v1 --source-bundle S3Bucket="$S3_BUCKET",S3Key="application.zip" --region $REGION || { echo "Failed to create application version. Exiting."; exit 1; }

# Create Elastic Beanstalk Environment if not exists
if aws elasticbeanstalk describe-environments --application-name $APP_NAME --environment-names $ENV_NAME | grep -q $ENV_NAME; then
    echo "Elastic Beanstalk environment $ENV_NAME already exists."
else
    echo "Creating Elastic Beanstalk environment $ENV_NAME..."
    aws elasticbeanstalk create-environment \
        --application-name $APP_NAME \
        --environment-name $ENV_NAME \
        --solution-stack-name "64bit Amazon Linux 2 v3.3.3 running Go 1.x" \  # Replace this with the valid solution stack
        --option-settings Namespace=aws:elasticbeanstalk:environment,OptionName=EnvironmentType,Value=LoadBalanced \
        --option-settings Namespace=aws:elasticbeanstalk:application:environment,OptionName=MONGO_URI,Value=${MONGO_URI} \
        --option-settings Namespace=aws:elasticbeanstalk:application:environment,OptionName=DB_NAME,Value=${DB_NAME} \
        --region $REGION || { echo "Failed to create Elastic Beanstalk environment. Exiting."; exit 1; }
fi

# Update environment to new version
echo "Updating environment to new version..."
aws elasticbeanstalk update-environment --environment-name $ENV_NAME --version-label v1 --region $REGION || { echo "Failed to update environment. Exiting."; exit 1; }

echo "Deployment to Elastic Beanstalk completed."
