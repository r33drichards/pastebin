## create an iam policy resource with full access to efs
resource "aws_iam_policy" "grafana" {
  name        = var.name
  description = "Full access to EFS"
  policy      = var.iam_policy
}

## create an iam role with the policy
resource "aws_iam_role" "grafana" {
  name               = var.name
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}
// attach the policy to the role
resource "aws_iam_role_policy_attachment" "grafana" {
  role       = "${aws_iam_role.grafana.name}"
  policy_arn = "${aws_iam_policy.grafana.arn}"
}

// create an instance profile
resource "aws_iam_instance_profile" "grafana" {
  name = var.name
  role = "${aws_iam_role.grafana.name}"
}
