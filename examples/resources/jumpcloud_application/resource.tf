resource "jumpcloud_application" "example" {
  name:"awsclientvpn-test",
  display_label:"AwsVpnTerraform-test",
  beta:false,
  sso_url:"https://sso.jumpcloud.com/saml2/awsclientvpn-test",
  learn_more:"https://support.jumpcloud.com/support/s/article/Single-Sign-On-SSO-with-AWS-Client-VPN",
  idp_certificate:"",
  idp_private_key:"",
  idp_entity_id:"JumpCloud",
  sp_entity_id:"urn:amazon:webservices:clientvpn",
  acs_url:"http://127.0.0.1:35001",
  constant_attributes:[],
}
