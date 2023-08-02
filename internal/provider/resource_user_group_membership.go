package provider

import (
	"context"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	jcapiv2 "github.com/TheJumpCloud/jcapi-go/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceUserGroupMembership() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource for managing user group memberships.",
		CreateContext: resourceUserGroupMembershipCreate,
		ReadContext:   resourceUserGroupMembershipRead,
		// We must not have an update routine as the association cannot be updated.
		// Any change in one of the elements forces a recreation of the resource
		Update:        nil,
		DeleteContext: resourceUserGroupMembershipDelete,
		Schema: map[string]*schema.Schema{
			"user_id": {
				Description: "The ID of the `resource_user` object.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"group_id": {
				Description: "The ID of the `resource_user_group` object.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: userGroupMembershipImporter,
		},
	}
}

// We cannot use the regular importer as it calls the read function ONLY with the ID field being
// populated.- In our case, we need the group ID and user ID to do the read - But since our
// artificial resource ID is simply the concatenation of user ID group ID seperated by  a '/',
// we can derive both values during our import process
func userGroupMembershipImporter(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	s := strings.Split(d.Id(), "/")
	_ = d.Set("group_id", s[0])
	_ = d.Set("user_id", s[1])
	return []*schema.ResourceData{d}, nil
}

func modifyUserGroupMembership(client *jcapiv2.APIClient,
	d *schema.ResourceData, action string) diag.Diagnostics {

	payload := jcapiv2.UserGroupMembersReq{
		Op:    action,
		Type_: "user",
		Id:    d.Get("user_id").(string),
	}

	req := map[string]interface{}{
		"body": payload,
	}

	_, err := client.UserGroupMembersMembershipApi.GraphUserGroupMembersPost(
		context.TODO(), d.Get("group_id").(string), "", "", req)

	return diag.FromErr(err)
}

func resourceUserGroupMembershipCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*jcapiv2.Configuration)
	client := jcapiv2.NewAPIClient(config)

	err := modifyUserGroupMembership(client, d, "add")
	if err != nil {
		return err
	}
	return resourceUserGroupMembershipRead(ctx, d, meta)
}

func resourceUserGroupMembershipRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*jcapiv2.Configuration)
	client := jcapiv2.NewAPIClient(config)

	for i := 0; i < 20; i++ { // Prevent infite loop
		optionals := map[string]interface{}{
			"group_id": d.Get("group_id").(string),
			"limit":    int32(100),
			"skip":     int32(i * 100),
		}

		graphconnect, _, err := client.UserGroupMembersMembershipApi.GraphUserGroupMembersList(
			context.TODO(), d.Get("group_id").(string), "", "", optionals)
		if err != nil {
			return diag.FromErr(err)
		}

		// The user_ids are hidden in a super-complex construct, see
		// https://github.com/TheJumpCloud/jcapi-go/blob/master/v2/docs/GraphConnection.md
		for _, v := range graphconnect {
			if v.To.Id == d.Get("user_id") {
				// Found - As we not have a JC-ID for the membership we simply store
				// the concatenation of group ID and user ID as our membership ID
				d.SetId(d.Get("group_id").(string) + "/" + d.Get("user_id").(string))
				return nil
			}
		}

		if len(graphconnect) < 100 {
			break
		} else {
			time.Sleep(100 * time.Millisecond)
		}
	}
	// Element does not exist in actual Infrastructure, hence unsetting the ID
	d.SetId("")
	return nil
}

func resourceUserGroupMembershipDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*jcapiv2.Configuration)
	client := jcapiv2.NewAPIClient(config)
	return modifyUserGroupMembership(client, d, "remove")
}
