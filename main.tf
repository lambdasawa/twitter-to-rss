terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.16"
    }
  }

  required_version = ">= 1.2.0"
}

provider "aws" {
  default_tags {
    tags = {
      repository = "github.com/lambdasawa/twitter-to-rss"
    }
  }
}

data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

resource "aws_dynamodb_table" "twitter_to_feed" {
  name = "TwitterToFeed"

  billing_mode = "PAY_PER_REQUEST"

  hash_key = "ID"

  attribute {
    name = "ID"
    type = "S"
  }
}

resource "aws_iam_role" "apprunner_access_role" {
  name = "twitter-to-rss-access-role"

  assume_role_policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Action" : "sts:AssumeRole",
        "Principal" : {
          "Service" : [
            "build.apprunner.amazonaws.com",
            "tasks.apprunner.amazonaws.com"
          ]
        },
        "Effect" : "Allow",
        "Sid" : ""
      }
    ]
  })
}

resource "aws_iam_policy" "apprunner_access_role" {
  name = "twitter-to-rss-apprunner-access-role"

  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        Action : [
          "ecr:GetDownloadUrlForLayer",
          "ecr:BatchGetImage",
          "ecr:DescribeImages",
          "ecr:GetAuthorizationToken",
          "ecr:BatchCheckLayerAvailability"
        ],
        Effect : "Allow",
        Resource : "*"
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "apprunner_access_role" {
  role       = aws_iam_role.apprunner_access_role.name
  policy_arn = aws_iam_policy.apprunner_access_role.arn
}

resource "aws_iam_role" "apprunner_instance_role" {
  name = "twitter-to-rss-apprunner-instnace-role"

  assume_role_policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Action" : "sts:AssumeRole",
        "Principal" : {
          "Service" : [
            "build.apprunner.amazonaws.com",
            "tasks.apprunner.amazonaws.com"
          ]
        },
        "Effect" : "Allow",
        "Sid" : ""
      }
    ]
  })
}

resource "aws_iam_policy" "apprunner_instance_role" {
  name = "twitter-to-rss-apprunner-instnace-role"

  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        Action : [
          "logs:CreateLogStream",
          "logs:PutLogEvents",
        ],
        Effect : "Allow",
        Resource : "*"
      },
      {
        Action : [
          "dynamodb:PutItem",
          "dynamodb:GetItem",
        ],
        Effect : "Allow",
        Resource : "*"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "main" {
  role       = aws_iam_role.apprunner_instance_role.name
  policy_arn = aws_iam_policy.apprunner_instance_role.arn
}

resource "aws_ecr_repository" "main" {
  name                 = "twitter-to-rss"
  image_tag_mutability = "MUTABLE"
}

data "archive_file" "main" {
  type        = "zip"
  source_dir  = "."
  output_path = "app.zip"
}

resource "null_resource" "ko_build" {
  triggers = {
    src_hash = data.archive_file.main.output_sha
  }

  provisioner "local-exec" {
    command = "ko build --bare --sbom=none ."

    working_dir = "."

    environment = {
      KO_DOCKER_REPO = "${data.aws_caller_identity.current.account_id}.dkr.ecr.${data.aws_region.current.name}.amazonaws.com/twitter-to-rss"
    }
  }

  depends_on = [data.archive_file.main, aws_ecr_repository.main]
}
resource "aws_apprunner_service" "main" {
  service_name = "twitter-to-rss"

  instance_configuration {
    instance_role_arn = aws_iam_role.apprunner_instance_role.arn
  }

  source_configuration {
    authentication_configuration {
      access_role_arn = aws_iam_role.apprunner_access_role.arn
    }
    image_repository {
      image_configuration {
        port = "8000"
      }
      image_identifier      = "${aws_ecr_repository.main.repository_url}:latest"
      image_repository_type = "ECR"
    }
  }

  health_check_configuration {
    protocol = "HTTP"
    path     = "/ping"
  }

  depends_on = [aws_ecr_repository.main, null_resource.ko_build]
}

output "service_url" {
  value = aws_apprunner_service.main.service_url
}
