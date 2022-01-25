package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	jcapiv1 "github.com/Paynetworx/jcapi-go/v1"
	jcapiv2 "github.com/TheJumpCloud/jcapi-go/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a JumpCloud system user resource. For additional information refer also to the [JumpCloud API user model](https://docs.jumpcloud.com/1.0/models/systemuserpost).",
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Schema: map[string]*schema.Schema{
			"username": {
				Description: "The technical user name. See JumpCloud's [user naming conventions](https://support.jumpcloud.com/support/s/article/naming-convention-for-users1) for naming restrictions. Example: `john.doe`.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"email": {
				Description: "The users e-mail address, which is also used for log ins. E-mail addresses have to be unique across all JumpCloud accounts, there cannot be two users with the same e-mail address. Example: `john.doe@acme.org`.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"firstname": {
				Description: "The user's first name. Example: `john`.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"lastname": {
				Description: "The user's last name. Example: `doe`.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"enable_mfa": {
				Description: "Require Multi-factor Authentication on the User Portal.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			// Currently, only the options necessary for our use case are implemented
			// JumpCloud offers a lot more
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	configv1 := convertV2toV1Config(meta.(*jcapiv2.Configuration))
	client := jcapiv1.NewAPIClient(configv1)

	payload := jcapiv1.Systemuserputpost{
		Username:                    d.Get("username").(string),
		Email:                       d.Get("email").(string),
		Firstname:                   d.Get("firstname").(string),
		Lastname:                    d.Get("lastname").(string),
		EnableUserPortalMultifactor: d.Get("enable_mfa").(bool),
	}
	req := map[string]interface{}{
		"body": payload,
	}
	returnStruct, _, err := client.SystemusersApi.SystemusersPost(context.TODO(),
		"", "", req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(returnStruct.Id)
	return resourceUserRead(ctx, d, meta)
}

func resourceUserRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	configv1 := convertV2toV1Config(meta.(*jcapiv2.Configuration))
	client := jcapiv1.NewAPIClient(configv1)

	res, _, err := client.SystemusersApi.SystemusersGet(context.TODO(),
		d.Id(), "", "", nil)

	// If the object does not exist in our infrastructure, we unset the ID
	// Unfortunately, the http request returns 200 even if the resource does not exist
	if err != nil {
		if err.Error() == "EOF" {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(res.Id)

	if err := d.Set("username", res.Username); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("email", res.Email); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("firstname", res.Firstname); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("lastname", res.Lastname); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("enable_mfa", res.EnableUserPortalMultifactor); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	configv1 := convertV2toV1Config(meta.(*jcapiv2.Configuration))
	client := jcapiv1.NewAPIClient(configv1)

	// The code from the create function is almost identical, but the structure is different :
	// jcapiv1.Systemuserput != jcapiv1.Systemuserputpost
	payload := jcapiv1.Systemuserput{
		Username:                    d.Get("username").(string),
		Email:                       d.Get("email").(string),
		Firstname:                   d.Get("firstname").(string),
		Lastname:                    d.Get("lastname").(string),
		EnableUserPortalMultifactor: d.Get("enable_mfa").(bool),
	}

	req := map[string]interface{}{
		"body": payload,
	}
	_, _, err := client.SystemusersApi.SystemusersPut(context.TODO(),
		d.Id(), "", "", req)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceUserRead(ctx, d, meta)
}

func resourceUserDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	configv1 := convertV2toV1Config(meta.(*jcapiv2.Configuration))
	client := jcapiv1.NewAPIClient(configv1)

	res, _, err := client.SystemusersApi.SystemusersDelete(context.TODO(),
		d.Id(), "", headerAccept, nil)
	if err != nil {
		// TODO: sort out error essentials
		return diag.Errorf("error deleting user group:%s; response = %+v", err, res)
	}
	d.SetId("")
	return nil
}
