
# M3
## Buckets
### Create Bucket
Required input:

1. Name `string`
2. Access `int`
	- *0* : Private 
	- *1* : Public
	- *2* : Custom

**Restrictions**:

  * Name should follow [S3 bucket naming convention](https://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-s3-bucket-naming-requirements.html).
  * If Global Bucket feature enabled, **bucket name** is unique accross tenants.

### Delete Bucket
Required input:

1. Name `string`

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
Required input:

1. Name `string`
2. Effect `"allow" | "deny"`  
3. Resources `[string]`
4. Action `"read" | "write" | "readwrite"`

**Restrictions**:

  * Permissions with the same name are not allowed.
  * Permissions with the exact same definition (effects, resources and actions) are not allowed.
  * A resource is a valid bucket name.
  * List of resources can't be empty.

### Update Permission
Updates the permissions with the desired definitions and applies the changes to the Service Accounts that are using it (if any).

Required input:

1. Id `string`
2. Name `string`
3. Effect `"allow" | "deny"`  
4. Resources `[string]`
5. Action `"read" | "write" | "readwrite"`

**Restrictions**:

  * List of resources can't be empty.

### Delete Permission
Deletes a permission and updates all the Service Accounts that are using it (if any).

Required input:

1. Id `string`

**Restrictions**:

  * If the permission to be deleted is the only one assigned to one or ore Service Accounts, the request will not be allowed.

## Service Accounts
### Create Service Account
Creates a Service Account and provides the `Access Key` and `Secret Key` to be used by the S3 application.

All permissions selected will be applied to the Service Account after its creation.
Required input:

1. Name `string`
2. Permissions Ids `[string]` 

**Restrictions**:

  * Service Accounts with the same name are not allowed.
  * List of permissions can't be empty.
  * **Secret Key** is not stored on the DB and is showed only after creation.

### Update Service Account
Updates the fields for a Service Account including the permissions assigned to it.

Required input:

1. Id `string`
2. Name `string`
3. Permissions Ids `[string]` 

**Restrictions**:

  * List of permissions can't be empty.

### Enable Service Account
Enables S3 requests for that Service Account.

Required input:

1. Id `string`

### Disable Service Account
Disables S3 requests for that Service Account.

Required input:

1. Id `string`

### Delete Service Account
Deletes a Service Account including its access.

Required input:

1. Id `string`

## Users
### Create User
Required input:

1. Name `string`
2. Email `string`

**Enabled** shows if the user is enabled or disabled -  `boolean`

### Enable User
Enables a user to login.

Required input:

1. Id `string`

### Disable User
Disables an user and invalidates all current active sessions.

Required input:

1. Id `string`

### Reset Password
Sends an email to an user to reset their password.

Required input:

1. Email `string`

**Restrictions**:

  * The email should belong to a valid registered user.

### Change Password
Resets an user's own password by retreiving the old and new password.

After changing password all user's current sessions get invalidated.
Required input:

1. Old Password `string`
2. New Password `string`

**Restrictions**:

  * New password is only set if a correct old password is provided.

### Forgot Password
Sends an email to an User to reset their password.

Required input:

1. Company `string`
2. Email `string`

**Restrictions**:

  * Any Company or Email will do a successful response.
  * Only if a valid Company and a valid User is provided on the request, the email will be sent.

