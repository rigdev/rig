import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Creating Groups

This document provides instructions on how to create groups using the SDK or CLI in Rig.

<hr class="solid" />

## Creating Groups

To create groups on your backend, you can utilize the `Create` endpoint. When making a request to create a group, it is necessary to set the `Name` field. This field specifies a unique name of the group being created:

<Tabs>
<TabItem value="go" label="Golang SDK">

```go
resp, err := client.Group().Create(ctx, connect.NewRequest(&group.CreateRequest{
  Initializers: []*group.Update{
    {Field: &group.Update_Name{Name: "admins"}},
  },
}))
if err != nil {
  log.Fatal(err)
}
log.Printf("successfully created group:\n%s\nWith ID: %s \n", resp.Msg.GetGroup().String(), resp.Msg.GetGroup().GetGroupId())
```

</TabItem>
<TabItem value="typescript" label="Typescript SDK">

```typescript
const resp = await client.groupsClient.create({
  initializers: [
    {
      field: {
        case: "name",
        value: "admins",
      },
    },
  ],
});
console.log(
  `successfully created group:\n${resp.group}\nWith ID: ${resp.group.name}`,
);
```

</TabItem>
<TabItem value="cli" label="CLI">

```sh
rig group create --name
```

The name is prompted if not provided by the flag
</TabItem>
</Tabs>

<hr class="solid" />

## Additional Fields

### Metadata

To add metadata to group requests, you can include one or multiple key-value pairs in the `Metadata` field. This allows you to attach additional information or custom data to the group:

<Tabs>
<TabItem value="go" label="Golang SDK">

```go
resp, err := client.Group().Create(ctx, connect.NewRequest(&group.CreateRequest{
  Initializers: []*group.Update{
    {Field: &group.Update_Name{Name: "admins"}},
    {Field: &group.Update_SetMetadata{SetMetadata: &model.Metadata{Key: "role", Value: []byte("1")}}},
  },
}))
if err != nil {
  log.Fatal(err)
}
log.Printf("successfully created group:\n%s\nWith ID: %s \n", resp.Msg.GetGroup().GetName(), resp.Msg.GetGroup().GetGroupId())
```

</TabItem>
<TabItem value="typescript" label="Typescript SDK">

```typescript
const resp = await client.groupsClient.create({
  initializers: [
    {
      field: {
        case: "name",
        value: "admins",
      },
    },
    {
      field: {
        case: "setMetadata",
        value: {
          key: "role",
          value: new TextEncoder().encode("1"),
        },
      },
    },
  ],
});
console.log(
  `successfully created group:\n${resp.group}\nWith ID: ${resp.group.name}`,
);
```

</TabItem>
<TabItem value="cli" label="CLI">

```sh
rig group create --name
rig group update [group-id | group-name] --field --value
```

Example:

```sh
rig group create admins
rig group update admins --field set-meta-data --value `{key:role,value:1}`
```

Setting these additional fields using the CLI requires first creating a group and then subsequently updating the group.
</TabItem>
</Tabs>
