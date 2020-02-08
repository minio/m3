
# M3
## Buckets
### Create Bucket
**Restrictions**:

  * Name should follow [S3 bucket naming convention](https://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-s3-bucket-naming-requirements.html).
  * If Global Bucket feature enabled, **bucket name** is unique accross tenants.
  * Only the Service Account with the proper permissions will be allowed to interact with the bucket.
### Delete Bucket
**Restrictions**:

  * Bucket can only be deleted if it is empty.
  * A Bucket can't be deleted if a permission is refering to it.

### List Bucket
Lists all buckets and usage [Bytes] per bucket.

Usage is calculated every 12 hrs.

> Edit is not allowed

## Permissions
A Permissions consists the following parts:

1. Name
2. Effect
	- Allow 
	- Deny
3. Resources 
4. Action
	- Read
	- Write
	- Read Write

M3 permissions is an abstraction and simplification of an S3 policy.
### Create Permission

**Restrictions**:

  * Permissions with the same name are not allowed.
  * Permissions with the exact same definition (effects, resources and actions) are not allowed.
  * A resource is a valid bucket name.
  * List of resources can't be empty.

### Update Permission
Updates the permissions with the desired definitions and applies the changes to the Service Accounts that are using it (if any).

**Restrictions**:

  * List of resources can't be empty.

### Delete Permission
Deletes a permission and updates all the Service Accounts that are using it (if any).

**Restrictions**:

  * If the permission to be deleted is the only one assigned to one or ore Service Accounts, the request will not be allowed.

## Service Accounts
### Create Service Account
Creates a Service Account and provides the `Access Key` and `Secret Key` to be used by the S3 application.

All permissions selected will be applied to the Service Account after its creation.

**Restrictions**:

  * Service Accounts with the same name are not allowed.
  * List of permissions can't be empty.
  * **Secret Key** is not stored on the DB and is showed only once during the time of creation.

### Update Service Account
Updates the fields for a Service Account including the permissions assigned to it.

**Restrictions**:

  * List of permissions can't be empty.

### Enable Service Account
Enables S3 requests for that Service Account.

### Disable Service Account
Disables S3 requests for that Service Account.

### Delete Service Account
Deletes a Service Account including its access.

## Users
### Create User
Creates a new user with the name and email.

**Enabled** shows if the user is enabled or disabled.

**Restrictions**:

  * Users with the same email are not allowed.

### Enable User
Enables a user to login.

### Disable User
Disables an user and invalidates all current active sessions.

### Reset Password
Sends an email to an user to reset their password.

**Restrictions**:

  * The email should belong to a valid registered user.

### Change Password
Resets an user's own password by retreiving the old and new password.

After changing password all user's current sessions get invalidated.

**Restrictions**:

  * New password is only set if a correct old password is provided.

### Forgot Password
Sends an email to an User to reset their password.

**Restrictions**:

  * Any Company or Email will do a successful response.
  * Only if a valid Company and a valid User is provided on the request, the email will be sent.

