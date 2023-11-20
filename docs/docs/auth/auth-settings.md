import InstallRig from '../../src/markdown/prerequisites/install-rig.md'
import SetupSdk from '../../src/markdown/prerequisites/setup-sdk.md'
import SetupCli from '../../src/markdown/prerequisites/setup-cli.md'
import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Managing Auth Settings

This document guides managing login settings and social providers using the SDK. You will learn how to configure and customize the available login options and integrate social providers into your application.

<hr class="solid" />

## Update Login Methods

To update the available login methods, you can utilize the `UpdateSettings` endpoint. The available login methods include email/password, phone/password, and username/password.

By default, the email/password and username/password flow is enabled. To modify the enabled login methods, make a call to `UpdateSettings` with your desired `LoginMechanisms`. In the example provided below, we enable all of the available login methods:

<Tabs>
<TabItem value="go" label="Golang SDK">

```go
if _, err := client.UserSettings().UpdateSettings(ctx, connect.NewRequest(&settings.UpdateSettingsRequest{
	Settings: []*settings.Update{
		{
			Field: &settings.Update_LoginMechanisms_{
				LoginMechanisms: &settings.Update_LoginMechanisms{
					LoginMechanisms: []model.LoginType{
						model.LoginType_LOGIN_TYPE_EMAIL_PASSWORD,
						model.LoginType_LOGIN_TYPE_USERNAME_PASSWORD,
						model.LoginType_LOGIN_TYPE_PHONE_PASSWORD,
					},
				},
			},
		},
	},
})); err != nil {
	log.Fatal(err)
}
```

</TabItem>
<TabItem value="typescript" label="Typescript SDK">

```typescript
await client.userSettings.updateSettings({
  settings: [
    {
      field: {
        case: "loginMechanisms",
        value: {
          loginMechanisms: [
            LoginType.EMAIL_PASSWORD,
            LoginType.USERNAME_PASSWORD,
            LoginType.PHONE_PASSWORD,
          ],
        },
      },
    },
  ],
});
```

</TabItem>
<TabItem value="cli" label="CLI">

```sh
rig user update-settings --field --value
```

Examples:

```sh
rig user update-settings -f login-mechanisms -v '[1,2,3]'
```

</TabItem>
</Tabs>

<hr class="solid" />

## Update Verification Settings

To update the verification settings for email and phone numbers, you can utilize the `UpdateSettings` endpoint. By default, both email and phone number verification are enabled (true). In the example below, we demonstrate how to disable email and phone number verification:

<Tabs>
<TabItem value="go" label="Golang SDK">

```go
if _, err := client.UserSettings().UpdateSettings(ctx, connect.NewRequest(&settings.UpdateSettingsRequest{
  Settings: []*settings.Update{
    {
      Field: &settings.Update_IsVerifiedEmailRequired{IsVerifiedEmailRequired: false},
    },
    {
      Field: &settings.Update_IsVerifiedPhoneRequired{IsVerifiedPhoneRequired: false},
    },
  },
})); err != nil {
  log.Fatal(err)
}
```

</TabItem>
<TabItem value="typescript" label="Typescript SDK">

```typescript
await client.userSettings.updateSettings({
  settings: [
    {
      field: { case: "isVerifiedEmailRequired", value: false },
    },
    {
      field: { case: "isVerifiedPhoneRequired", value: false },
    },
  ],
});
```

</TabItem>
<TabItem value="cli" label="CLI">
```sh
rig user update-settings --field --value
```

Example:

```sh
rig user update-settings -f verify-email-required -v false
rig user update-settings -f verify-phone-required -v false
```

</TabItem>
</Tabs>

<hr class="solid" />

## Configure OAuth Providers

To configure your OAuth providers, you can make use of the `UpdateSettings` endpoint. Before proceeding with the configuration, ensure that you have created the necessary OAuth apps for each provider. The available providers and their respective setup guides are as follows:

- Google: [Create Google OAuth App](https://support.google.com/cloud/answer/6158849?hl=en)
- GitHub: [Create GitHub OAuth App](https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/creating-an-oauth-app)
- Facebook: [Create Facebook OAuth App](https://developers.facebook.com/docs/development/create-an-app/)

After setting up your desired provider, you can use the `UpdateSettings` endpoint with the following parameters to complete the configuration:

<Tabs>
<TabItem value="go" label="Golang SDK">

```go
if _, err := client.UserSettings().UpdateSettings(ctx, connect.NewRequest(&settings.UpdateSettingsRequest{
    Settings: []*settings.Update{
        {
          Field: &settings.settings.Update_CallbackUrls_{
              CallbackUrls: &settings.Update_CallbackUrls{
                  Callbacks: []string{
                      "http://localhost:3000/callback"
                  },
              },
          },
        },
        {
            Field: &settings.Update_OauthProvider{
                OauthProvider: &settings.OauthProviderUpdate{
                    Provider: model.OauthProvider_OAUTH_PROVIDER_GOOGLE,
                    Credentials: &model.ProviderCredentials{
                        PublicKey:   "GOOGLE-OAUTH-CLIENT-ID",
                        PrivateKey:  "GOOGLE-OAUTH-SECRET",
                    },
                    AllowLogin:    true,
                    AllowRegister: true,
                },
            },
        },
    },
})); err != nil {
    log.Fatal(err)
}
```

</TabItem>
<TabItem value="typescript" label="Typescript SDK">

```typescript
await client.userSettings.updateSettings({
  settings: [
    {
      field: {
        case: "callbackUrls",
        value: { callbackUrls: ["http://localhost:3000/callback"] },
      },
    },
    {
      field: {
        case: "oauthProvider",
        value: {
          provider: OAuthProvider.GOOGLE,
          credentials: {
            publicKey: "GOOGLE-OAUTH-CLIENT-ID",
            privateKey: "GOOGLE-OAUTH-SECRET",
          },
          allowLogin: true,
          allowRegister: true,
        },
      },
    },
  ],
});
```

</TabItem>
<TabItem value="cli" label="CLI">

```sh
rig user update-settings -f callbacks -v '["rig.dev", "rig.org"]'
rig user update-settings -f oauth-settings -v '{"provider": "OAUTH_PROVIDER_GOOGLE", "credentials": {"public_key": "GOOGLE-OAUTH-CLIENT-ID", "private_key": "GOOGLE-OAUTH-SECRET"}, "allow_login": true, "allow_register": true}'
```

</TabItem>
</Tabs>

The `Callbacks` field is essential as it declares the endpoints to which Rig is permitted to redirect users after they have logged in. These endpoints can be either your own HTTP server, where you can retrieve the access and refresh tokens or your frontend application.

If you are utilizing a provider other than `Google`, simply replace `Google` with `Facebook` or `GitHub` according to your chosen provider.


## Update TTL on Codes
You can manage the validity duration of access tokens, refresh tokens, and verification codes. This functionality can be valuable for limiting token usage for enhanced security or extending session persistence for longer durations. You can customize these durations based on your specific requirements and preferences.

To manage the TTL, you can utilize the `UpdateSettings` endpoint as below:

<Tabs>
<TabItem value="go" label="Golang SDK">

```go
if _, err := client.UserSettings().UpdateSettings(ctx, connect.NewRequest(&settings.UpdateSettingsRequest{
    Settings: []*settings.Update{
        {
            Field: &settings.Update_AccessTokenTtl{AccessTokenTtl: durationpb.New(time.Second * 60 * 5)}, // set to 5 minutes
        },
        {
            Field: &settings.Update_RefreshTokenTtl{RefreshTokenTtl: durationpb.New(time.Hour * 24)}, // set to 24 hours
        },
        {
            Field: &settings.Update_VerificationCodeTtl{VerificationCodeTtl: durationpb.New(time.Second * 60 * 10)}, // set to 10 minutes
        },
    },
})); err != nil {
    log.Fatal(err)
}
```

</TabItem>
<TabItem value="typescript" label="Typescript SDK">

```typescript
await client.userSettings.updateSettings({
  settings: [
    {
      field: {
        case: "accessTokenTtl",
        value: new Duration({ seconds: 5 * 60 }),
      },
    },
    {
      field: {
        case: "refreshTokenTtl",
        value: new Duration({ seconds: 24 * 60 * 60 }),
      },
    },
    {
      field: {
        case: "verificationCodeTtl",
        value: new Duration({ seconds: 10 * 60 }),
      },
    },
  ],
});
```

</TabItem>
<TabItem value="cli" label="CLI">

In the example provided, the access tokens have a validity of 5 minutes, refresh tokens are valid for 24 hours, and verification codes remain valid for 10 minutes.

It is important to note that when refreshing tokens after the access token has expired, a new pair of refresh and access tokens will be generated with an extended validity period.

```sh
rig user update-settings --field --value
```

Example:

```sh
rig user update-settings -f access-token-ttl -v 5
rig user update-settings -f refresh-token-ttl -v 24
rig user update-settings -f verification-code-ttl -v 10
```

</TabItem>
</Tabs>

<hr class="solid" />

## Update Hashing Algorithm

To update the password hashing algorithm, you can make use of the `UpdateSettings` endpoint. This functionality can be beneficial for security purposes or if you are migrating users from a different provider. By default, the hashing algorithm is set to `Bcrypt` with a cost of `10`. In the example provided below, we illustrate how to update the BCrypt cost to `12`:

<Tabs>
<TabItem value="go" label="Golang SDK">

```go
if _, err := client.UserSettings().UpdateSettings(ctx, connect.NewRequest(&settings.UpdateSettingsRequest{
    Settings: []*settings.Update{
        {
            Field: &settings.Update_PasswordHashing{PasswordHashing: &model.HashingConfig{
                Method: &model.HashingConfig_Bcrypt{Bcrypt: &model.BcryptHashingConfig{
                    Cost: 12,
                }},
            }},
        },
    },
})); err != nil {
    log.Fatal(err)
}
```

</TabItem>
<TabItem value="typescript" label="Typescript SDK">

```typescript
await client.userSettings.updateSettings({
  settings: [
    {
      field: {
        case: "passwordHashing",
        value: {
          method: {
            case: "bcrypt",
            value: { cost: 12 },
          },
        },
      },
    },
  ],
});
```

</TabItem>
<TabItem value="cli" label="CLI">

```sh
rig user update-settings --field --value
```

Example:

```sh
rig user update-settings -f password-hashing -v '{"bcrypt": {"cost": 12}}'
```

</TabItem>
</Tabs>

