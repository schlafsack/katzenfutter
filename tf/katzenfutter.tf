// https://aws.amazon.com/blogs/startups/from-zero-to-eks-with-terraform-and-helm/

locals {
  aws_region = "eu-west-2"
  aws_profile = "greasley"
  aws_credentials = "~/.aws/credentials"
  application = "katzenfutter"
  common_tags = {
    appplication:local.application
  }
  cluster_azs = [
    "eu-west-2a",
    "eu-west-2b",
    "eu-west-2c",
  ]
}

data "aws_eks_cluster" "cluster" {
  name = module.katzenfutter_cluster.cluster_id
}

data "aws_eks_cluster_auth" "cluster" {
  name = module.katzenfutter_cluster.cluster_id
}

provider "kubernetes" {
  host = data.aws_eks_cluster.cluster.endpoint
  cluster_ca_certificate = base64decode(data.aws_eks_cluster.cluster.certificate_authority.0.data)
  token = data.aws_eks_cluster_auth.cluster.token
  load_config_file = false
  version = ">=1.11.1"
}

provider "aws" {
  version = ">=2.52.0"
  region = local.aws_region
  profile = local.aws_profile
  shared_credentials_file = local.aws_credentials
}

provider "helm" {
  version = ">=1.0.0"
  debug = true
  kubernetes {
    host = data.aws_eks_cluster.cluster.endpoint
    cluster_ca_certificate = base64decode(data.aws_eks_cluster.cluster.certificate_authority.0.data)
    token = data.aws_eks_cluster_auth.cluster.token
    load_config_file = false
  }
}

module "katzenfutter_vpc" {
  # https://github.com/terraform-aws-modules/terraform-aws-vpc
  source = "terraform-aws-modules/vpc/aws"
  name = "${local.application}-vpc"
  cidr = "10.0.0.0/16"
  azs = local.cluster_azs
  enable_nat_gateway = true
  single_nat_gateway = true
  tags = merge(local.common_tags, {
    "kubernetes.io/cluster/${local.application}" = "shared"
  })
  private_subnet_tags = merge(local.common_tags, {
    "kubernetes.io/cluster/${local.application}" = "shared",
    "kubernetes.io/role/elb" = "1"
  })
  public_subnet_tags = merge(local.common_tags, {
    "kubernetes.io/cluster/${local.application}" = "shared"
    "kubernetes.io/role/elb" = "1"
  })
  private_subnets = [
    "10.0.1.0/24",
    "10.0.2.0/24",
    "10.0.3.0/24",
  ]
  public_subnets = [
    "10.0.4.0/24",
    "10.0.5.0/24",
    "10.0.6.0/24",
  ]
}

module "katzenfutter_cluster" {
  # https://github.com/terraform-aws-modules/terraform-aws-eks
  source = "terraform-aws-modules/eks/aws"
  cluster_name = local.application

  subnets = module.katzenfutter_vpc.private_subnets
  vpc_id = module.katzenfutter_vpc.vpc_id
  write_kubeconfig = true

  node_groups = {
    (local.application) = {
      name = "${local.application}-default-nodegroup"
      desired_capacity = 3
      max_capacity = 10
      min_capacity = 1
      instance_type = "m4.large"
      additional_tags = merge(local.common_tags, {})
    }
  }

  tags = merge(local.common_tags, {})
}

resource "kubernetes_namespace" "katzenfutter_zeebe_namespace" {
  metadata {
    name = "${local.application}-zeebe"
  }
}

data "helm_repository" "zeebe_repository" {
  name = "zeebe"
  url = "https://helm.zeebe.io"
}

resource "helm_release" "katzenfutter_zeebe" {
  name = "katzenfutter-zeebe"
  chart = "zeebe/zeebe-full"
  repository = data.helm_repository.zeebe_repository.name
  namespace = kubernetes_namespace.katzenfutter_zeebe_namespace.metadata[0].name
}
