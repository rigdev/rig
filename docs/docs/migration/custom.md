import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Migrate To Rig
This document guides how to migrate from other systems to Rig. Rig provides a general framework for migrating users, as well as implementations of this framework for specific systems such as Firebase. The concept of migrating users, and the authentication of these, revolve around migrating the users' credentials and the specific hashing parameters of these to Rig. Thus to migrate users to Rig, it is necessary to provide the hashing parameters of the source system as well as the hashed credentials of the users.

## Migrate Users Endpoint

In case you would like to migrate from a service that Rig does not yet provide a migration from, you can implement your own migration tool, by creating users with the hashed credentials and hashed parameters. This is done using the `CreateUsers` endpoint. In the example below, we migrate a single user from a toy system:

<Tabs>
<TabItem value="go" label="Golang SDK">

```go
userPasswordHash := "YourUserPasswordHash"
userPasswordSalt := "YourUserPasswordSalt"

hashingConfig := &model.HashingConfig{
    Method: &model.HashingConfig_Scrypt{
        Scrypt: &model.ScryptHashingConfig{
            SignerKey:     "YourSignerKey",
            SaltSeparator: "YourSaltSeparator",
            Rounds:        8,
            MemCost:       14,
            P:             1,
            KeyLen:        int32(32),
        },
    },
}

hashingInstance := &model.HashingInstance{
    Config: hashingConfig,
    Hash: userPasswordHash,
    Instance: &model.HashingInstance_Scrypt{
        Scrypt: &model.ScryptHashingInstance{
            Salt: userPasswordSalt,
        },
    },
}

initializers := []*user.Update{
    {
        Field: &user.Update_HashedPassword{
            HashedPassword: hashingInstance,
        },
    },
    {
        Field: &user.Update_Email{
            Email: "john@doe.com",
        },
    },
    {
        Field: &user.Update_PhoneNumber{
            PhoneNumber: "+4588888888"
        }
    }
}

_, err := client.User().Create(ctx, &connect.NewRequest(&user.CreateRequest{
    Initializers: initializers
}))
if err != nil {
    log.Fatal(err)
}

log.Printf("Successfully migrated user")
```

</TabItem>
</Tabs>
