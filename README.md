# awswhois

CLI tool to identify which AWS service and region an IP address or hostname belongs to.

## Installation

```bash
go install github.com/maelvls/awswhois@latest
```

## Usage

```bash
# Check an IP address
awswhois 3.4.12.4

# Check a hostname
awswhois api-dev210.qa.venafi.io
```

## Example Output

```
$ awswhois 3.4.12.4
IP        PREFIX       REGION     SERVICE  BORDER GROUP
3.4.12.4  3.4.12.4/32  eu-west-1  AMAZON   eu-west-1

$ awswhois api-dev210.qa.venafi.io
IP             PREFIX         REGION     SERVICE     BORDER GROUP
54.200.113.36  54.200.0.0/15  us-west-2  AMAZON,EC2  us-west-2
35.155.136.3   35.155.0.0/16  us-west-2  AMAZON,EC2  us-west-2
54.149.89.45   54.148.0.0/15  us-west-2  AMAZON,EC2  us-west-2

$ awswhois s3.amazonaws.com
IP              PREFIX          REGION     SERVICE        BORDER GROUP
16.182.68.128   16.182.0.0/16   us-east-1  AMAZON,S3      us-east-1
52.217.161.88   52.216.0.0/15   us-east-1  AMAZON,S3      us-east-1
16.15.195.248   16.15.192.0/18  us-east-1  AMAZON,S3,EC2  us-east-1
...
```

## How It Works

1. Fetches the latest AWS IP ranges from https://ip-ranges.amazonaws.com/ip-ranges.json
2. Resolves hostnames to IP addresses (supports both IPv4 and IPv6)
3. Checks each IP against all AWS CIDR ranges
4. Groups results by IP prefix, region, and border group
5. Displays matching AWS services (comma-separated if multiple), region, and network border group information
