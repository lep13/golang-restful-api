#!/bin/bash

# Make sure the reload script is executable
chmod +x .platform/hooks/postdeploy/01_reload_nginx.sh

# Test the nginx configuration to ensure it's valid
sudo nginx -t
if [ $? -eq 0 ]; then
    # Reload nginx if the configuration test passes
    sudo service nginx reload
else
    echo "Nginx configuration is invalid, not reloading."
    exit 1
fi
