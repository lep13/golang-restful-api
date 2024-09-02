#!/bin/bash

# Set AWS region and application/environment names
AWS_REGION="us-east-1"
APP_NAME="golang-restful-api"
ENV_NAME="golang-restful-api-env"

# Create Elastic Beanstalk application if it doesn't exist
aws elasticbeanstalk describe-applications --application-names $APP_NAME --region $AWS_REGION | grep "ApplicationName"
if [ $? -ne 0 ]; then
    aws elasticbeanstalk create-application --application-name $APP_NAME --region $AWS_REGION
fi

# Create environment with necessary configuration
aws elasticbeanstalk create-environment --application-name $APP_NAME --environment-name $ENV_NAME \
    --solution-stack-name "64bit Amazon Linux 2 v5.4.6 running Go 1.x" \
    --option-settings Namespace=aws:elasticbeanstalk:application:environment,OptionName=MONGO_URI,Value=$MONGO_URI \
    --option-settings Namespace=aws:elasticbeanstalk:application:environment,OptionName=DB_NAME,Value=$DB_NAME \
    --region $AWS_REGION

# Wait for environment to be ready
aws elasticbeanstalk wait environment-exists --application-name $APP_NAME --environment-name $ENV_NAME --region $AWS_REGION

# Deploy the application
aws elasticbeanstalk update-environment --application-name $APP_NAME --environment-name $ENV_NAME --version-label latest --region $AWS_REGION

# Configure CloudWatch Logs
aws elasticbeanstalk create-configuration-template --application-name $APP_NAME --template-name CWLogsConfig \
    --option-settings Namespace=aws:elasticbeanstalk:hostmanager,OptionName=LogPublicationControl,Value=true \
    --region $AWS_REGION
