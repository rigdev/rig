import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Creating Users

This document provides instructions on how to create users using the SDK or the CLI in Rig. Users in Rig can be created with one or more of three types of unique identifiers:

- `Email` - You can use any valid and unique [email address](https://en.wikipedia.org/wiki/Email_address).
- `Phone number` - You can use any valid and unique [phone number](https://en.wikipedia.org/wiki/E.164) as an identifier for creating users.
- `Username` - You can use any unique username without spaces as an identifier.

To create users, you can utilize the `Create` endpoint. This endpoint takes a list of initializers that specify the fields to be set on the user.

<hr class="solid" />

## Email and Password

Set the `Email` field to specify the desired email for the user, and use the `Password` field to assign a password for the user:

<Tabs>
<TabItem value="go" label="Golang SDK">

```go
// create a user with email/password
resp, err := client.User().Create(ctx, connect.NewRequest(&user.CreateRequest{
  Initializers: []*user.Update{
    {
      Field: &user.Update_Email{Email: "john@acme.com"},
    },
    {
      Field: &user.Update_Password{Password: "TestPassword1234!"},
    },
  },
}))
if err != nil {
    log.Fatal(err)
}
log.Printf("successfully created user: %s \n", res.Msg.GetUser().GetUserInfo().GetEmail())
log.Printf("with userID: %s \n", resp.Msg.GetUser().GetUserId())
```

</TabItem>
<TabItem value="typescript" label="Typescript SDK">

```typescript
// create a user with email/password
const resp = await client.user.create({
  initializers: [
    {
      field: {
        case: "email",
        value: "john@acme.com",
      },
    },
    {
      field: {
        case: "password",
        value: "TestPassword1234!",
      },
    },
  ],
});
console.log(`successfully created user: ${resp.user?.userInfo?.email}`);
console.log(`with userID: ${resp.user?.userId}`);
```

</TabItem>
<TabItem value="cli" label="CLI">

```sh
rig user create --email --username --phone --password
```

Example:

```sh
rig user create -e john@acme.com -p TestPassword1234!
```

</TabItem>
</Tabs>

## Phone Number and Password

Set the `PhoneNumber` field to specify the desired phone number for the user, and use the `Password` field to assign a password for the user:

<Tabs>
<TabItem value="go" label="Golang SDK">

```go
// create a user with phone/password
resp, err := client.User().Create(ctx, connect.NewRequest(&user.CreateRequest{
  Initializers: []*user.Update{
    {
      Field: &user.Update_PhoneNumber{PhoneNumber: "+4522122798"},
    },
    {
      Field: &user.Update_Password{Password: "TestPassword1234!"},
    },
  },
}))
if err != nil {
    log.Fatal(err)
}
log.Printf("successfully created user: %s \n", resp.Msg.GetUser().GetUserInfo().GetPhoneNumber())
log.Printf("with userID: %s \n", resp.Msg.GetUser().GetUserId())
```

</TabItem>
<TabItem value="typescript" label="Typescript SDK">

```typescript
// create a user with phone/password
const resp = await client.user.create({
  initializers: [
    {
      field: {
        case: "phoneNumber",
        value: "+4522122798",
      },
    },
    {
      field: {
        case: "password",
        value: "TestPassword1234!",
      },
    },
  ],
});
console.log(`successfully created user: ${resp.user?.userInfo?.phoneNumber}`);
console.log(`with userID: ${resp.user?.userId}`);
```

</TabItem>
<TabItem value="cli" label="CLI">

```sh
rig user create --email --username --phone --password
```

Example:

```sh
rig user create -p +4522122798 -p TestPassword1234!
```

</TabItem>
</Tabs>

## Username and Password

To create users on your backend, you can utilize the `Create` endpoint. Set the `Username` field to specify the desired username for the user, and use the `Password` field to assign a password for the user:

<Tabs>
<TabItem value="go" label="Golang SDK">

```go
// create a user with username/password
resp, err := client.User().Create(ctx, connect.NewRequest(&user.CreateRequest{
  Initializers: []*user.Update{
    {
      Field: &user.Update_Username{Username: "markGrayson1234"},
    },
    {
      Field: &user.Update_Password{Password: "TestPassword1234!"},
    },
  },
}))
if err != nil {
    log.Fatal(err)
}
log.Printf("successfully created user: %s \n", resp.Msg.GetUser().GetUserInfo().GetUsername())
log.Printf("with userID: %s \n", resp.Msg.GetUser().GetUserId())
```

</TabItem>
<TabItem value="typescript" label="Typescript SDK">

```typescript
// create a user with username/password
const resp = await client.user.create({
  initializers: [
    {
      field: {
        case: "username",
        value: "markGrayson1234",
      },
    },
    {
      field: {
        case: "password",
        value: "TestPassword1234!",
      },
    },
  ],
});
console.log(`successfully created user: ${resp.user?.userInfo?.username}`);
console.log(`with userID: ${resp.user?.userId}`);
```

</TabItem>
<TabItem value="cli" label="CLI">

```sh
rig user create --email --username --phone --password
```

Example:

```sh
rig user create -u markGrayson1234 -p TestPassword1234!
```

</TabItem>
</Tabs>

## Additional Fields

### Profile Information

When creating users using the SDK, you have the option to include additional profile information. The following profile information fields are available:

- `FirstName` - This field represents the first name of the user.
- `LastName` - This field represents the last name of the user.
- `PreferredLanguage` - This field indicates the preferred language of the user. Any string value is valid.
- `Country` - This field represents the user's country of origin. Any string value is valid.

Set the `Profile` field as below as part of creating your users:

<Tabs>
<TabItem value="go" label="Golang SDK">

```go
// create a user with profile information
profile := &user.Update_Profile{
  Profile: &user.Profile{
    FirstName:         "John",
    LastName:          "Doe",
    PreferredLanguage: "DK",
    Country:           "Denmark",
  },
}
resp, err := client.User().Create(ctx, connect.NewRequest(&user.CreateRequest{
  Initializers: []*user.Update{
    {
      Field: &user.Update_Username{Username: "markGrayson1234"},
    },
    {
      Field: &user.Update_Password{Password: "TestPassword1234!"},
    },
    {
      Field: profile,
    },
  },
}))
if err != nil {
    log.Fatal(err)
}
log.Printf("successfully created user: %s \n", resp.Msg.GetUser().GetUserInfo().GetUsername())
log.Printf("with userID: %s \n", resp.Msg.GetUser().GetUserId())
```

</TabItem>
<TabItem value="typescript" label="Typescript SDK">

```typescript
// create a user with profile information
const resp = await client.user.create({
  initializers: [
    {
      field: {
        case: "username",
        value: "markGrayson1234",
      },
    },
    {
      field: {
        case: "password",
        value: "TestPassword1234!",
      },
    },
    {
      field: {
        case: "profile",
        value: {
          firstName: "John",
          lastName: "Doe",
          preferredLanguage: "DK",
          country: "Denmark",
        },
      },
    },
  ],
});
console.log(`successfully created user: ${resp.user?.userInfo?.username}`);
console.log(`with userID: ${resp.user?.userId}`);
```

</TabItem>

<TabItem value="cli" label="CLI">

```sh
rig user update [user-id | {email|username|phone}] --field --value
```

Example:

```sh
rig user update markGrayson1234 -f profile -v '{"firstName":"John","lastName":"Doe","preferredLanguage":"DK","country":"Denmark"}'
```

Setting these additional fields using the CLI requires first creating a user and then subsequently updating the user where the updates are prompted.
</TabItem>
</Tabs>

### Metadata

As part of the user creation process, you can include custom metadata for your users. This metadata is added as a byte array. In the following example, we will utilize JSON marshalling to convert our data into bytes and store it on the user:

<Tabs>
<TabItem value="go" label="Golang SDK">

```go
type Account struct {
	Age      int
	StripeId string
}
...
// example metadata
metadata, err := json.Marshal(&Account{
  Age:      18,
  StripeId: "user-stripe-id",
})
if err != nil {
    log.Fatal(err)
}
// create a user with metadata
resp, err := client.User().Create(ctx, connect.NewRequest(&user.CreateRequest{
  Initializers: []*user.Update{
    {
      Field: &user.Update_Username{Username: "markGrayson1234"},
    },
    {
      Field: &user.Update_Password{Password: "TestPassword1234!"},
    },
    {
      Field: &user.Update_SetMetadata{
        SetMetadata: &model.Metadata{Key: "account", Value: metadata},
      },
    },
    {
      Field: &user.Update_SetMetadata{
        SetMetadata: &model.Metadata{Key: "color", Value: []byte("blue")},
      },
    },
  },
}))
if err != nil {
    log.Fatal(err)
}
log.Printf("successfully created user: %s \n", resp.Msg.GetUser().GetUserInfo().GetUsername())
log.Printf("with userID: %s \n", resp.Msg.GetUser().GetUserId())
```

</TabItem>
<TabItem value="typescript" label="Typescript SDK">

```typescript
// create a user with metadata
const account = { age: 18, stripeId: "user-stripe-id" };
const resp = await client.user.create({
  initializers: [
    {
      field: {
        case: "username",
        value: "markGrayson1234",
      },
    },
    {
      field: {
        case: "password",
        value: "TestPassword1234!",
      },
    },
    {
      field: {
        case: "setMetadata",
        value: {
          key: "account",
          value: new TextEncoder().encode(JSON.stringify(account)),
        },
      },
    },
    {
      field: {
        case: "setMetadata",
        value: {
          key: "color",
          value: new TextEncoder().encode("blue"),
        },
      },
    },
  ],
});
console.log(`successfully created user: ${resp.user?.userInfo?.username}`);
console.log(`with userID: ${resp.user?.userId}`);
```

</TabItem>

<TabItem value="cli" label="CLI">

```sh
rig user update [user-id | {email|username|phone}]
```

Example:

```sh
rig user update markGrayson1234 -f set-meta-data -v '{"key":"account","value":{"age":18,"stripeId":"user-stripe-id"}}'
rig user update markGrayson1234 -f set-meta-data -v '{"key":"color","value":"blue"}'
```

Setting these additional fields using the CLI requires first creating a user and then subsequently updating the user.
</TabItem>
</Tabs>

Notice that we are inserting two metadata keys in the above example: `account` and `color`, by creating two `SetMetadata` fields.
