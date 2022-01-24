package provider

import (
	jcapiv1 "github.com/TheJumpCloud/jcapi-go/v1"
	jcapiv2 "github.com/TheJumpCloud/jcapi-go/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/net/context"
	"log"
)

type Constant struct {
	name string
	value string
	read_only bool
	required bool
	visible bool
}

func resourceApplication() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource for adding an Amazon Web Services (AWS) account application. **Note:** This resource is due to change in future versions to be more generic and allow for adding various applications supported by JumpCloud.",
		CreateContext: resourceApplicationCreate,
		ReadContext:   resourceApplicationRead,
		UpdateContext: resourceApplicationUpdate,
		DeleteContext: resourceApplicationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the application",
				Type:        schema.TypeString,
				Required:    true,
			},
			"beta": {
				Description: "",
				Type:        schema.TypeBool,
				Default:	 false,
				Optional:      true,
			},
			"display_label": {
				Description: "Name of the application to display",
				Type:        schema.TypeString,
				Required:    true,
			},
			"sso_url": {
				Description: "The SSO URL suffix to use",
				Type:        schema.TypeString,
				Required:    true,
			},
			"learn_more": {
				Description: "",
				Type:        schema.TypeString,
				Optional:	 true,
			},
			"constant_attributes":{
				Description:	"",
				Type:			schema.TypeList,
				Optional:	    true,
				Elem:			&schema.Resource{
					Schema: map[string]*schema.Schema{
						"name":{
							Type: schema.TypeString,
							Required: true,
						},
						"value":{
							Type: schema.TypeString,
							Required: true,
						},
						"read_only":{
							Type: schema.TypeBool,
							Optional: true,
							Default: false,
						},
						"required":{
							Type: schema.TypeBool,
							Optional: true,
							Default: false,
						},
						"visible":{
							Type: schema.TypeBool,
							Optional: true,
							Default: true,
						},
					},
				},
			},
			"idp_certificate":{
				Description: "",
				Type:        schema.TypeString,
				Required:    true,
			},
			"idp_entity_id":{
				Description: "",
				Type:        schema.TypeString,
				Required:    true,
			},
			"idp_private_key":{
				Description: "",
				Type:        schema.TypeString,
				Required:    true,
			},
			"sp_entity_id":{
				Description: "",
				Type:        schema.TypeString,
				Required:    true,
			},
			"acs_url":{
				Description: "",
				Type:        schema.TypeString,
				Required:    true,
			},
			"metadata_xml": {
				Description: "The JumpCloud metadata XML file.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceApplicationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	configv1 := convertV2toV1Config(meta.(*jcapiv2.Configuration))
	client := jcapiv1.NewAPIClient(configv1)

	payload := generateApplicationPayload(d)
	request := map[string]interface{}{
		"body": payload,
	}

	returnStruct, _, err := client.ApplicationsApi.ApplicationsPost(context.TODO(), request)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(returnStruct.Id)
	return resourceApplicationRead(ctx, d, meta)
}

func resourceApplicationRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	configv1 := convertV2toV1Config(meta.(*jcapiv2.Configuration))
	client := jcapiv1.NewAPIClient(configv1)

	response, _, err := client.ApplicationsApi.ApplicationsGet(context.TODO(), d.Id(), nil)

	// If the object does not exist, unset the ID
	if err != nil {
		if err.Error() == "EOF" {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(response.Id)

	if err := d.Set("display_label", response.DisplayLabel); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("sso_url", response.SsoUrl); err != nil {
		return diag.FromErr(err)
	}

	constantAttributeValues := response.Config.ConstantAttributes.Value
	constants := []Constant{}
	for _, el := range constantAttributeValues {
		constant := Constant{}	
		constant.name = el.Name
		constant.read_only = el.ReadOnly
		constant.value = el.Value
		constant.required = el.Required
		constant.visible = el.Visible 
		constants = append(constants, constant)
	}
	d.Set("constant_attributes",constants)
	d.Set("acs_url",response.Config.AcsUrl.Value)
	d.Set("idp_certificate",response.Config.IdpCertificate.Value)
	d.Set("idp_private_key", response.Config.IdpPrivateKey.Value)
	d.Set("idp_entity_id", response.Config.IdpEntityId.Value)
	d.Set("sp_entity_id", response.Config.SpEntityId.Value)
	d.Set("name",response.Name)
	d.Set("beta",response.Beta)

	if response.Id != "" {
		log.Println("[INFO] response ID is ", response.Id)
		orgId := configv1.DefaultHeader["x-org-id"]
		apiKey := configv1.DefaultHeader["x-api-key"]

		metadataXml, err := GetApplicationMetadataXml(orgId, response.Id, apiKey)
		if err != nil {
			return diag.FromErr(err)
		}

		if err := d.Set("metadata_xml", metadataXml); err != nil {
			return diag.FromErr(err)
		}
	} else {
		log.Println("[INFO] no ID in response, skipping metadata XML retrieval")
	}

	return nil
}

func resourceApplicationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	configv1 := convertV2toV1Config(meta.(*jcapiv2.Configuration))
	client := jcapiv1.NewAPIClient(configv1)

	payload := generateApplicationPayload(d)
	request := map[string]interface{}{
		"body": payload,
	}

	_, _, err := client.ApplicationsApi.ApplicationsPut(context.TODO(), d.Id(), request)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceApplicationRead(ctx, d, meta)
}

func resourceApplicationDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	configv1 := convertV2toV1Config(meta.(*jcapiv2.Configuration))
	client := jcapiv1.NewAPIClient(configv1)

	_, _, err := client.ApplicationsApi.ApplicationsDelete(context.TODO(), d.Id(), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func generateApplicationPayload(d *schema.ResourceData) jcapiv1.Application {
	constants := []jcapiv1.ApplicationConfigConstantAttributesValue{}
	for _, data := range d.Get("constant_attributes").([]schema.ResourceData) {
		constant := jcapiv1.ApplicationConfigConstantAttributesValue{}
		constant.Name = data.Get("name").(string)
		constant.Value = data.Get("value").(string)
		constant.ReadOnly= data.Get("read_only").(bool)
		constant.Required = data.Get("required").(bool)
		constant.Visible = data.Get("visible").(bool)
		constants = append(constants, constant)
	}
	return jcapiv1.Application{
		// TODO clearify if previous Active: true is translated to Beta: false
		Beta:         d.Get("beta").(bool),
		Name:         d.Get("name").(string),
		DisplayLabel: d.Get("display_label").(string),
		SsoUrl:       d.Get("sso_url").(string),
		Config: &jcapiv1.ApplicationConfig{
			AcsUrl: &jcapiv1.ApplicationConfigAcsUrl{
				Type_:"text",
				Label:"ACS Url:",
				Value:d.Get("acs_url").(string),
				Required:true,
				Visible:true,
				ReadOnly:false,
				Position:4,
			},
			ConstantAttributes: &jcapiv1.ApplicationConfigConstantAttributes{
				Value: constants,
			},
			DatabaseAttributes: &jcapiv1.ApplicationConfigDatabaseAttributes{},
			IdpCertificate:&jcapiv1.ApplicationConfigAcsUrl{
				Type_:"file",
				Label:"IdP Certificate:",
				Value:d.Get("idp_certificate").(string),
				Required:true,
				Visible:true,
				ReadOnly:false,
				Position:2,
			},
			IdpEntityId:&jcapiv1.ApplicationConfigAcsUrl{
				Type_:"text",
				Label:"IdP Entity ID:",
				Value:d.Get("idp_entity_id").(string),
				Required:true,
				Visible:true,
				ReadOnly:false,
				Position:0,
			},
			IdpPrivateKey:&jcapiv1.ApplicationConfigAcsUrl{
				Type_:"file",
				Label:"IdP Private Key:",
				Value:d.Get("idp_private_key").(string),
				Required:true,
				Visible:true,
				ReadOnly:false,
				Position:1,
			},
			SpEntityId:&jcapiv1.ApplicationConfigAcsUrl{
				Type_:"text",
				Label:"SP Entity ID:",
				Value:d.Get("sp_entity_id").(string),
				Required:true,
				Visible:true,
				ReadOnly:false,
				Position:4,
			},
		},
	}
}
