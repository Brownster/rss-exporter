package collectors

import "testing"

func TestParseAWSGUID_Basic(t *testing.T) {
	svc, region := ParseAWSGUID("https://status.aws.amazon.com/#athena-us-west-2_1234")
	if svc != "athena" || region != "us-west-2" {
		t.Fatalf("got %s %s", svc, region)
	}
}

func TestParseAWSGUID_ARN(t *testing.T) {
	svc, region := ParseAWSGUID("arn:aws:health:us-east-1::event/AWS_EC2_EXAMPLE")
	if svc != "ec2" || region != "us-east-1" {
		t.Fatalf("got %s %s", svc, region)
	}
}
