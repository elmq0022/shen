# SHEN

## Status

In Development

## Description

A token based authorization service that facilitates Role-Based Access Control (RBAC), written in go.

## Goals

- Provide a simple but usable authorization system for managing access primarily for internal systems
- Provide username and password based auth for real users
- Provide token based auth for real users and service accounts
- Administrators can register applications with Shenlong
- Administrators can create user groups
- Users can belong to one or more user groups
- User groups can be mapped to one or more combinations of application and RBAC role 
- None is a valid RBAC role and implies that the application does not use RBAC or manages RBAC itself
- shen provides a CLI for administration of the above. 
