#!/bin/bash

# Variables from environment
APP_NAME="golang-restful-api"
ENV_NAME=${EB_ENV_NAME}
S3_BUCKET=${S3_BUCKET_NAME}
REGION=${AWS_REGION}

# Create S3 bucket if not exists
if aws s3api head-bucket --bucket "$S3_BUCKET" 2>/dev/null; then
    echo "S3 bucket $S3_BUCKET already exists."
else
    echo "Creating S3 bucket $S3_BUCKET..."
    aws s3 mb s3://$S3_BUCKET --region $REGION
fi

# Create Elastic Beanstalk Application if not exists
if aws elasticbeanstalk describe-applications --application-names $APP_NAME | grep -q $APP_NAME; then
    echo "Elastic Beanstalk application $APP_NAME already exists."
else
    echo "Creating Elastic Beanstalk application $APP_NAME..."
    aws elasticbeanstalk create-application --application-name $APP_NAME --region $REGION
fi

# Create Elastic Beanstalk Environment if not exists
if aws elasticbeanstalk describe-environments --application-name $APP_NAME --environment-names $ENV_NAME | grep -q $ENV_NAME; then
    echo "Elastic Beanstalk environment $ENV_NAME already exists."
else
    echo "Creating Elastic Beanstalk environment $ENV_NAME..."
    aws elasticbeanstalk create-environment \
        --application-name $APP_NAME \
        --environment-name $ENV_NAME \
        --solution-stack-name "64bit Amazon Linux 2 v3.1.3 running Go 1.x" \
        --option-settings Namespace=aws:elasticbeanstalk:environment,OptionName=EnvironmentType,Value=LoadBalanced \
        --option-settings Namespace=aws:elasticbeanstalk:application:environment,OptionName=MONGO_URI,Value=${MONGO_URI} \
        --option-settings Namespace=aws:elasticbeanstalk:application:environment,OptionName=DB_NAME,Value=${DB_NAME} \
        --region $REGION
fi

# Package application
zip -r application.zip .

# Upload package to S3
aws s3 cp application.zip s3://$S3_BUCKET/application.zip

# Create new application version
aws elasticbeanstalk create-application-version --application-name $APP_NAME --version-label v1 --source-bundle S3Bucket="$S3_BUCKET",S3Key="application.zip" --region $REGION

# Update environment to new version
aws elasticbeanstalk update-environment --environment-name $ENV_NAME --version-label v1 --region $REGION

echo "Deployment to Elastic Beanstalk completed."
