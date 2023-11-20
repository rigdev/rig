---
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Managing Users using the SDK or CLI

This document provides instructions on managing users using the SDK or CLI. It covers various operations such as fetching users and sessions, as well as updating user profiles and users' contact information.

<hr class="solid" />

## Fetching Users

### Getting a User

To retrieve users in Rig, you can utilize the `Get` endpoint by including the unique `UUID` of the user in your request, or you can use the `Lookup` endpoint by including a unique identifier such as an email address, username, or phone number:

<Tabs>
<TabItem value="go" label="Golang SDK">

```go
userID := "" // NOTE: insert a specifc userID here
resp, err := client.User().Get(ctx, connect.NewRequest(&user.GetRequest{
    UserId: userID,
}))
if err != nil {
    log.Fatal(err)
}
log.Printf("successfully fetched user: %s \n", resp.Msg.GetUser().GetUserId())

// OR

identifier := &model.UserIdentifier{} // NOTE: insert a specifc identifier here
resp, _ := client.User().Lookup(ctx, connect.NewRequest(&user.LookupRequest{
    Identifier: identifier,
}))
log.Printf("successfully fetched user: %s \n", resp.Msg.GetUser().GetUserId())
```

</TabItem>
<TabItem value="typescript" label="Typescript SDK">

```typescript
const userID = "" // NOTE: insert a specifc userID here
const resp = await client.users.get({
    userId: userID,
})
console.log(`successfully fetched user: ${resp.user}`)

// OR

const identifier := {} // NOTE: insert a specifc identifier here
const resp = await client.users.lookup({
    identifier: identifier,
})
console.log(`successfully fetched user: ${resp.user}`)
```

</TabItem>
<TabItem value="cli" label="CLI">

```sh
rig user get [user-id | {email|username|phone}] --json
```

Example:

```sh
rig user get john@acme.com
```

</TabItem>
</Tabs>

### Listing Users

To list users in Rig, you can utilize the `List` endpoint. Pagination can be implemented using the `Offset` and `Limit` fields. The `Offset` field determines the starting point of the list, while the `Limit` field specifies the maximum number of users to retrieve in each request.

<Tabs>
<TabItem value="go" label="Golang SDK">

```go
resp, err := client.User().List(ctx, connect.NewRequest(&user.ListRequest{
    Pagination: &model.Pagination{
        Offset: 10,
        Limit:  10,
    },
}))
if err != nil {
    log.Fatal(err)
}
log.Printf("successfully fetched %d users. Total amount is: %d \n", len(resp.Msg.GetUsers()), resp.Msg.GetTotal())
```

</TabItem>
<TabItem value="typescript" label="Typescript SDK">

```typescript
const resp = await client.user.list({
  pagination: {
    offset: 10,
    limit: 10,
  },
});
console.log(
  `successfully fetched ${resp.users.length} users. Total amount is: ${resp.total}`,
);
```

</TabItem>

<TabItem value="cli" label="CLI">

```sh
rig user list [search...] --offset --limit --json
```

Example:

```sh
rig user list -o 10 -l 10 --json
```

</TabItem>
</Tabs>

The above query will fetch `10` users in your project, starting from user `10`. The total amount of users is returned as well.

### Searching for Users

In Rig, you can search for users by adding the `Search` parameter to your `List` request. All users in Rig have indexes on the following fields:

- Email addresses
- Phone numbers
- Usernames
- First names
- Last names

Here's an example where we search for users named "John":

<Tabs>
<TabItem value="go" label="Golang SDK">

```go
resp, err := client.User().List(ctx, connect.NewRequest(&user.ListRequest{
    Pagination: &model.Pagination{
        Offset: 10,
        Limit:  10,
    },
    Search: "john"
}))
if err != nil {
    log.Fatal(err)
}
log.Printf("successfully fetched %d client.User() matching your query. Total amount is: %d \n", len(resp.Msg.GetUsers()), resp.Msg.GetTotal())
```

</TabItem>
<TabItem value="typescript" label="Typescript SDK">

```typescript
const resp = await client.user.list({
  pagination: {
    offset: 10,
    limit: 10,
  },
  search: "john",
});
console.log(
  `successfully fetched ${resp.users.length} users matching your query. Total amount is: ${resp.total}`,
);
```

</TabItem>
<TabItem value="cli" label="CLI">

```sh
rig user list [search...] --offset --limit --json
```

Example:

```sh
rig user list john -o 10 -l 10
```

</TabItem>
</Tabs>

<hr class="solid" />

## Updating Users

To update users in Rig, you can make use of the `Update` endpoint. It is possible to include one or multiple updates in your request to modify multiple fields simultaneously.

Here's an example:

<Tabs>
<TabItem value="go" label="Golang SDK">

```go
userID := "" // NOTE: insert a specifc userID here
if _, err := client.User().Update(ctx, connect.NewRequest(&user.UpdateRequest{
    UserId: userID,
    Updates: []*user.Update{
        {   // To update email
            Field: &user.Update_Email{Email: "mark@acme.com"},
        },
        {   // To phone number
            Field: &user.Update_PhoneNumber{PhoneNumber: "+4588888888"},
        },
        {   // To update username
            Field: &user.Update_Username{Username: "MarkGrayson1234"},
        },
        {   // To update password
            Field: &user.Update_Password{Password: "Password123!"},
        },
        {   // To update profile
            Field: &user.Update_Profile{Profile: &user.Profile{
                FirstName: "Mark",
                LastName:  "Grayson",
            }},
        },
        {   // To update email verification status
            Field: &user.Update_IsEmailVerified{IsEmailVerified: true},
        },
        {   // To update phone verification status
            Field: &user.Update_IsPhoneVerified{IsPhoneVerified: true},
        },
        {
            Field: &user.Update_ResetSessions_{
                ResetSessions: &user.Update_ResetSessions{},
            },
        },
        {
            Field: &user.Update_SetMetaData{
                SetMetaData: &model.Metadata{
                    Key:   "superhero",
                    Value: []byte("1"),
                }
            }
        },
        {
            Field: &user.Update_DeleteMetaData{DeleteMetaDataKey: "from_earth"}
        }
    },
})); err != nil {
    log.Fatal(err)
}
log.Println("successfully updated user")
```

</TabItem>
<TabItem value="typescript" label="Typescript SDK">

```typescript
const userID = ""; // NOTE: insert a specifc userID here
await client.user.update({
  userId: userID,
  updates: [
    {
      // To update email
      field: {
        case: "email",
        value: "mark@acme.com",
      },
    },
    {
      // To phone number
      field: {
        case: "phoneNumber",
        value: "+4588888888",
      },
    },
    {
      // To update username
      field: {
        case: "username",
        value: "MarkGrayson1234",
      },
    },
    {
      // To update password
      field: {
        case: "password",
        value: "Password123!",
      },
    },
    {
      // To update profile
      field: {
        case: "profile",
        value: {
          firstName: "Mark",
          lastName: "Grayson",
        },
      },
    },
    {
      // To update email verification status
      field: {
        case: "isEmailVerified",
        value: true,
      },
    },
    {
      // To update phone verification status
      field: {
        case: "isPhoneVerified",
        value: true,
      },
    },
    {
      field: {
        case: "resetSessions",
        value: {},
      },
    },
    {
      field: {
        case: "setMetadata",
        value: {
          key: "superhero",
          value: new TextEncoder().encode("1"),
        },
      },
    },
    {
      field: {
        case: "deleteMetadata",
        value: "from_earth",
      },
    },
  ],
});
console.log(`successfully updated user`);
```

</TabItem>
<TabItem value="cli" label="CLI">

```sh
rig user update [user-id | {email|username|phone}] --field --value
```

Example:

```sh
rig user update john@acme.com -f email -v mark@acme.com
rig user update mark@acme.com -f phone-number -v +4588888888
rig user update +4588888888 -f username -v MarkGrayson1234
rig user update MarkGrayson1234 -f password -v TeamRig23!
rig user update MarkGrayson1234 -f profile -v '{"first_name":"Mark","last_name":"Grayson"}'
rig user update MarkGrayson1234 -f email-verified -v true
rig user update MarkGrayson1234 -f phone-verified -v true
rig user update MarkGrayson1234 -f reset-sessions -v '{}' # NOTE: the value is ignored
rig user update MarkGrayson1234 -f set-metadata -v '{"key": "superhero", "value": 1}' # NOTE: Only string meta-data can be using the CLI
rig user update MarkGrayson1234 -f delete-metadata -v from_earth
```

When specifying a field- and value-pair, only a single field can be updated at a time. If no field- and value-pair is specified, the command will be interactive, and many fields can be updated at a time

</TabItem>
</Tabs>

**Please notice that when updating a `profile`, all profile fields **will be **overridden**, also** empty fields (fields not set).**

<hr class="solid" />

## List Sessions for a User

In Rig, a user session represents a successful login flow. To retrieve a list of sessions for a specific user, you can utilize the `ListSessions` endpoint. By making a request to this endpoint, you can access the session information associated with the user's login activities.

<Tabs>
<TabItem value="go" label="Golang SDK">

```go
const userID := "" // NOTE: insert a specifc userID here
resp, err := client.User().ListSessions(ctx, connect.NewRequest(&user.ListSessionsRequest{
    UserId: userID,
}))
if err != nil {
    log.Fatal(err)
}
log.Printf("successfully fetched %d user sessions \n", resp.Msg.GetTotal())
```

</TabItem>
<TabItem value="typescript" label="Typescript SDK">

```typescript
const userID = ""; // NOTE: insert a specifc userID here
const resp = await client.user.listSessions({
  userId: userID,
});
console.log(`successfully fetched ${resp.total} user sessions`);
```

</TabItem>
<TabItem value="cli" label="CLI">

```sh
rig user list-sessions [user-id | {email|username|phone}] --json
```

Example:

```sh
rig user list-sessions mark@acme.com --json
```

</TabItem>
</Tabs>
