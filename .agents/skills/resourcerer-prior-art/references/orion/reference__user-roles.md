# User roles

> The four Orion user roles and what each can do


Orion uses role-based access control. Each user in your organization is assigned one of the following roles.

| Role | Permissions |
|------|-------------|
| **Viewer** | View optimization data and export recommendations |
| **Operator** | Apply and export recommendations; configure workloads |
| **Manager** | Invite users (not suspend); all Operator permissions |
| **Administrator** | Invite and suspend users; configure global settings |
{.td-initial .table}

Operators, Managers, and Administrators can also add and revoke Access Credentials under **Settings > Access Credentials**.

## Role comparison

| Action | Viewer | Operator | Manager | Administrator |
|--------|:------:|:--------:|:-------:|:-------------:|
| View optimization data | ✓ | ✓ | ✓ | ✓ |
| Export recommendations | ✓ | ✓ | ✓ | ✓ |
| Apply recommendations | | ✓ | ✓ | ✓ |
| Configure workloads | | ✓ | ✓ | ✓ |
| Add/revoke Access Credentials | | ✓ | ✓ | ✓ |
| Invite users | | | ✓ | ✓ |
| Suspend users | | | | ✓ |
| Configure global settings | | | | ✓ |
{.td-initial .table}
