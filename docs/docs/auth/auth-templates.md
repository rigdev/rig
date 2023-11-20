import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Update Templates
To update a template, you have to use the `UpdateSettings` endpoint within `ProjectSettings`. You need to pass a subject for the email (the name of the email you want to send) together with a body in HTML or plaintext. You can update the following templates using the SDK:

- `Welcome Email` - The email sent to a user when they either sign up or are created by an admin.
- `Reset Password Email` - The email sent to a user when they request to reset their password.
- `Verify Email` - The email sent to a user, when they login with their account for the first time using email, and is required to verify their email.

Below is an example where we update the email verification email:

<Tabs>
<TabItem value="go" label="Golang SDK">

```go
_, err := client.ProjectSettings().UpdateSettings(ctx, connect.NewRequest(&settings.UpdateSettingsRequest{
    Updates: []*settings.Update{
        {Field: &settings.Update_Templates{
            Templates: &settings.Templates{
                VerifyEmail: &settings.Template{
                    Body:    "<h1>Please verify your email {{ .Identifier }} using the following code: {{ .Code }}</h1>",
                    Subject: "",
                },
            },
        }},
    },
}))
if err != nil {
    log.Fatal(err)
}
log.Println("successfully updated email template")
```

</TabItem>
<TabItem value="typescript" label="Typescript SDK">

```typescript
await client.projectSettings.updateSettings({
    updates: [
        {
            field: {
                case: "templates",
                value: {
                    body: "<h1>Please verify your email {{ .Identifier }} using the following code: {{ .Code }}</h1>",
                    subject: "",
                    }
                }
        }},
    ],
})
console.log("successfully updated email template")
```

</TabItem>
<TabItem value="cli" label="CLI">

```sh
rig project update-settings --field --value
```

Examples:

```sh
rig project update-settings -f template -v '{"type": "TEMPLATE_TYPE_EMAIL_VERIFICATION", "body": "<h1>Please verify your email {{ .Identifier }} using the following code: {{ .Code }}</h1>", "subject": ""}'
```

</TabItem>
</Tabs>
