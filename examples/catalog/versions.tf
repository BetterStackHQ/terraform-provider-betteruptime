terraform {
  required_version = ">= 0.15"
  required_providers {
    betteruptime = {
      source  = "BetterStackHQ/better-uptime"
      version = ">= 0.2.4"
    }
  }
}