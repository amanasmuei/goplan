# GoPlan Administrator Guide

## Overview

This guide is for GoPlan administrators responsible for managing users, teams, and system configuration. Administrators have elevated privileges to configure the platform for their organization.

## Admin Access

### Accessing Admin Panel

1. Log in to GoPlan with an admin account
2. Click your profile icon in the top-right
3. Select **"Admin Panel"** from the dropdown

The Admin Panel provides access to:
- User Management
- Team Management
- System Settings
- Analytics & Reports
- Audit Logs

### Admin Roles

| Role | Capabilities |
|------|-------------|
| **Super Admin** | Full access to all settings and data |
| **Team Admin** | Manage users and settings within assigned teams |
| **Support Admin** | Read-only access for troubleshooting |

## User Management

### Adding Users

**Single User:**
1. Go to **Admin Panel > Users**
2. Click **"Add User"**
3. Enter user details:
   - Email address
   - Name
   - Role (User, Team Lead, Admin)
   - Team assignment
4. Click **"Send Invitation"**

The user will receive an email invitation to set up their account.

**Bulk Import:**
1. Go to **Admin Panel > Users > Import**
2. Download the CSV template
3. Fill in user details
4. Upload the completed CSV
5. Review and confirm the import

### Managing User Permissions

1. Go to **Admin Panel > Users**
2. Click on a user
3. Select the **"Permissions"** tab
4. Modify permissions as needed:
   - View own tasks only
   - View team tasks
   - Create tasks
   - Assign tasks to others
   - Access analytics
   - Admin capabilities

### Deactivating Users

1. Go to **Admin Panel > Users**
2. Find the user and click **"..."** menu
3. Select **"Deactivate"**
4. Confirm deactivation

Deactivated users:
- Cannot log in
- Tasks remain in the system
- Historical data is preserved
- Can be reactivated later

### Password Reset

For SSO-integrated deployments:
- Direct users to your identity provider's password reset

For standalone authentication:
1. Go to **Admin Panel > Users**
2. Select the user
3. Click **"Reset Password"**
4. User will receive a reset email

## Team Management

### Creating Teams

1. Go to **Admin Panel > Teams**
2. Click **"Create Team"**
3. Enter team details:
   - Team name
   - Description
   - Team lead
   - Department (optional)
4. Click **"Create"**

### Assigning Users to Teams

1. Go to **Admin Panel > Teams**
2. Select a team
3. Click **"Add Members"**
4. Search and select users
5. Click **"Add"**

Users can belong to multiple teams. Their primary team determines their default dashboard view.

### Team Settings

Each team can configure:
- Default task priorities
- Notification preferences
- Working hours (for due date calculations)
- Task templates

## System Configuration

### General Settings

Navigate to **Admin Panel > Settings > General**:

| Setting | Description |
|---------|-------------|
| Organization Name | Displayed in header and emails |
| Default Timezone | Used when user hasn't set preference |
| Working Hours | Mon-Fri 9am-5pm by default |
| Weekend Tasks | Allow/disallow weekend due dates |

### Authentication Settings

**SSO Configuration:**
1. Go to **Settings > Authentication**
2. Enable SSO
3. Configure your identity provider:
   - SAML 2.0 (Okta, Azure AD, etc.)
   - OAuth 2.0 (Google, GitHub, etc.)
4. Enter provider details (Entity ID, SSO URL, Certificate)
5. Test the connection
6. Enable for all users

**Session Settings:**
- Session timeout (default: 24 hours)
- Remember me duration (default: 30 days)
- Concurrent session limit

### Email Configuration

Configure SMTP for notifications:

```
SMTP Host: smtp.yourcompany.com
SMTP Port: 587
Username: goplan@yourcompany.com
Password: ********
From Address: goplan@yourcompany.com
From Name: GoPlan
```

Test configuration with **"Send Test Email"**.

### AI Settings

Configure AI prediction behavior:

| Setting | Description | Default |
|---------|-------------|---------|
| Prediction Confidence | Minimum confidence to show prediction | 60% |
| Learning Rate | How quickly AI adapts to feedback | Medium |
| Historical Window | Days of data used for predictions | 90 |
| Minimum Data Points | Tasks needed before predictions start | 10 |

## Data Management

### Importing Data

**From CSV:**
1. Go to **Admin Panel > Data > Import**
2. Select **"Tasks"** or **"Users"**
3. Download and fill the template
4. Upload the CSV
5. Map columns to fields
6. Review and import

**From Other Tools:**
GoPlan supports direct import from:
- Jira (via API)
- Asana (via export)
- Trello (via JSON export)

### Exporting Data

1. Go to **Admin Panel > Data > Export**
2. Select data type (Tasks, Users, Analytics)
3. Set date range and filters
4. Choose format (CSV, JSON, PDF)
5. Click **"Generate Export"**

Large exports are processed in the background. You'll receive an email when ready.

### Data Retention

Configure retention policies:

| Data Type | Default Retention | Options |
|-----------|------------------|---------|
| Active Tasks | Forever | N/A |
| Completed Tasks | 2 years | 1-5 years |
| Cancelled Tasks | 1 year | 6 months - 2 years |
| Audit Logs | 1 year | 1-3 years |
| Analytics | 2 years | 1-5 years |

### Backup & Recovery

GoPlan automatically backs up data:
- **Daily backups**: Retained for 30 days
- **Weekly backups**: Retained for 12 weeks
- **Monthly backups**: Retained for 12 months

To request a data recovery:
1. Contact GoPlan support
2. Specify the date/time to restore
3. Confirm the recovery scope

## Monitoring & Analytics

### System Health Dashboard

The health dashboard shows:
- Active users (real-time)
- API response times
- Error rates
- Storage usage
- AI prediction accuracy

### Usage Reports

Generate reports on:
- Active users by period
- Tasks created/completed
- Feature adoption
- Team productivity comparisons

### Audit Logs

All admin actions are logged:
1. Go to **Admin Panel > Audit Logs**
2. Filter by:
   - User
   - Action type
   - Date range
3. Export for compliance

Logged actions include:
- User creation/modification
- Permission changes
- Settings changes
- Data exports
- Login attempts

## API Administration

### API Keys

Manage API access:

1. Go to **Admin Panel > API**
2. Click **"Create API Key"**
3. Set:
   - Name (for identification)
   - Permissions (read, write, admin)
   - Expiration
4. Copy the generated key (shown only once)

### Rate Limits

Default API limits:
- **Standard**: 1000 requests/hour
- **Elevated**: 10000 requests/hour
- **Enterprise**: Unlimited

Adjust per API key in the API settings.

### Webhooks

Configure webhooks for integrations:

1. Go to **Admin Panel > Webhooks**
2. Click **"Add Webhook"**
3. Configure:
   - URL endpoint
   - Events to trigger (task.created, task.completed, etc.)
   - Secret for verification
4. Test and save

## Troubleshooting

### Common Issues

**Users Can't Log In:**
1. Verify SSO configuration
2. Check user status (not deactivated)
3. Review audit logs for failed attempts
4. Test with a known working account

**Slow Performance:**
1. Check system health dashboard
2. Review current user count
3. Check for bulk operations in progress
4. Contact support if persistent

**Missing Data:**
1. Verify user permissions
2. Check filters and date ranges
3. Review audit logs for deletions
4. Request data recovery if needed

### Support Escalation

For issues requiring GoPlan support:

1. Gather information:
   - Error messages (screenshots)
   - Steps to reproduce
   - Affected users/teams
   - Time of occurrence

2. Check system status at status.goplan.com

3. Contact support:
   - Email: support@goplan.com
   - Include your organization ID

## Security Best Practices

1. **Enable SSO** - Centralize authentication
2. **Regular audits** - Review user permissions quarterly
3. **Minimum permissions** - Grant only necessary access
4. **API key rotation** - Rotate keys every 90 days
5. **Monitor logs** - Set up alerts for suspicious activity
6. **Data exports** - Review and restrict export permissions

## Appendix: Permission Matrix

| Action | User | Team Lead | Admin |
|--------|:----:|:---------:|:-----:|
| View own tasks | ✓ | ✓ | ✓ |
| Create tasks | ✓ | ✓ | ✓ |
| Assign to self | ✓ | ✓ | ✓ |
| Assign to others | | ✓ | ✓ |
| View team tasks | | ✓ | ✓ |
| View team analytics | | ✓ | ✓ |
| Manage team members | | ✓ | ✓ |
| View all tasks | | | ✓ |
| Manage users | | | ✓ |
| System settings | | | ✓ |
| Audit logs | | | ✓ |
| API management | | | ✓ |

---

*GoPlan Admin Guide v1.0*
