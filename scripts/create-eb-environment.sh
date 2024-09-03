#!/bin/bash

# Variables from environment
APP_NAME="golang-restful-api"
ENV_NAME=${EB_ENV_NAME}
S3_BUCKET=${S3_BUCKET_NAME}
REGION=${AWS_REGION}

# Check required environment variables
if [ -z "$S3_BUCKET" ]; then
    echo "Error: S3_BUCKET_NAME is not set. Exiting."
    exit 1
fi

if [ -z "$ENV_NAME" ]; then
    echo "Error: EB_ENV_NAME is not set. Exiting."
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

# Create Elastic Beanstalk Application if not exists
if aws elasticbeanstalk describe-applications --application-names $APP_NAME | grep -q $APP_NAME; then
    echo "Elastic Beanstalk application $APP_NAME already exists."
else
    echo "Creating Elastic Beanstalk application $APP_NAME..."
    aws elasticbeanstalk create-application --application-name $APP_NAME --region $REGION
fi

# Create new application version
echo "Creating new application version..."
aws elasticbeanstalk create-application-version --application-name $APP_NAME --version-label v1 --source-bundle S3Bucket="$S3_BUCKET",S3Key="application.zip" --region $REGION

# Create Elastic Beanstalk Environment if not exists
if aws elasticbeanstalk describe-environments --application-name $APP_NAME --environment-names $ENV_NAME | grep -q $ENV_NAME; then
    echo "Elastic Beanstalk environment $ENV_NAME already exists."
else
    echo "Creating Elastic Beanstalk environment $ENV_NAME..."
    aws elasticbeanstalk create-environment \
    --application-name $APP_NAME \
    --environment-name $ENV_NAME \
    --solution-stack-name "64bit Amazon Linux 2 v3.11.0 running Go 1" \
    --option-settings file://option-settings.json \
    --option-settings Namespace=aws:autoscaling:launchconfiguration,OptionName=IamInstanceProfile,Value=ElasticBeanstalk-InstanceProfile \
    --region $REGION
fi

echo "Deployment to Elastic Beanstalk completed."
