# SHENLONG

## Status

In Development

## Description

A token based authorization service supporting roles based access control (RBAC), written in go.

## Goals

- Provide a simple but useable authorization system for managing access primarly for internal systems
- Provide username and password base auth for real users
- Provide token based auth for real users and service accounts
- Administrators can register applications with Shenlong
- Administrators can create user groups
- Users can belong to one or more user groups
- User groups can be mapped to one or more combinations of application and RBAC role 
- None is a valid RBAC role and implies that application does not use RBAC or mananages RBAC itself
- Shenlong provides a CLI for administation of the above. 
