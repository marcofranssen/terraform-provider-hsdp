package hsdp

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/philips-software/go-hsdp-api/cartel"
	"time"
)

func resourceContainerHost() *schema.Resource {
	return &schema.Resource{
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Create: resourceContainerHostCreate,
		Read:   resourceContainerHostRead,
		Update: resourceContainerHostUpdate,
		Delete: resourceContainerHostDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(12 * time.Minute),
			Update: schema.DefaultTimeout(12 * time.Minute),
			Delete: schema.DefaultTimeout(22 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"instance_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "m5.large",
			},
			"volume_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"iops": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"protect": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"encrypt_volumes": {
				Type:     schema.TypeBool,
				Default:  true,
				Optional: true,
				ForceNew: true,
			},
			"volumes": {
				Type:     schema.TypeInt,
				Default:  0,
				Optional: true,
				ForceNew: true,
			},
			"volume_size": {
				Type:     schema.TypeInt,
				Default:  0,
				Optional: true,
				ForceNew: true,
			},
			"security_groups": {
				Type:     schema.TypeSet,
				MaxItems: 5,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"user_groups": {
				Type:     schema.TypeSet,
				MaxItems: 50,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"private_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"role": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"launch_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"block_devices": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceContainerHostCreate(d *schema.ResourceData, m interface{}) error {
	config := m.(*Config)
	client, err := config.CartelClient()
	if err != nil {
		return err
	}

	tagName := d.Get("name").(string)
	protect := d.Get("protect").(bool)
	iops := d.Get("iops").(int)
	encryptVolumes := d.Get("encrypt_volumes").(bool)
	volumeSize := d.Get("volume_size").(int)
	numberOfVolumes := d.Get("volumes").(int)
	volumeType := d.Get("volume_type").(string)
	instanceType := d.Get("instance_type").(string)
	securityGroups := expandStringList(d.Get("security_groups").(*schema.Set).List())
	userGroups := expandStringList(d.Get("user_groups").(*schema.Set).List())

	ch, _, err := client.Create(tagName,
		cartel.SecurityGroups(securityGroups...),
		cartel.UserGroups(userGroups...),
		cartel.VolumeType(volumeType),
		cartel.IOPs(iops),
		cartel.InstanceType(instanceType),
		cartel.VolumesAndSize(numberOfVolumes, volumeSize),
		cartel.VolumeEncryption(encryptVolumes),
		cartel.Protect(protect),
	)
	if err != nil {
		return err
	}
	d.SetId(ch.InstanceID())

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"pending"},
		Target:     []string{"running"},
		Refresh:    InstanceStateRefreshFunc(client, tagName, []string{"terminated", "shutting-down"}),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf(
			"Error waiting for instance (%s) to become ready: %s",
			ch.InstanceID(), err)
	}
	d.SetConnInfo(map[string]string{
		"type": "ssh",
		"host": ch.IPAddress(),
	})
	return nil
}

func resourceContainerHostDelete(d *schema.ResourceData, m interface{}) error {
	config := m.(*Config)
	client, err := config.CartelClient()
	if err != nil {
		return err
	}

	tagName := d.Get("name").(string)
	ch, _, err := client.GetDetails(tagName)
	if err != nil {
		return err
	}
	if ch.InstanceID != d.Id() {
		return ErrInstanceIDMismatch
	}
	_, _, err = client.Destroy(tagName)
	if err != nil {
		return err
	}
	d.SetId("")
	return nil

}

func resourceContainerHostUpdate(d *schema.ResourceData, m interface{}) error {
	config := m.(*Config)
	client, err := config.CartelClient()
	if err != nil {
		return err
	}

	tagName := d.Get("name").(string)
	ch, _, err := client.GetDetails(tagName)
	if err != nil {
		return err
	}
	if ch.InstanceID != d.Id() {
		return ErrInstanceIDMismatch
	}

	d.Partial(true)

	if d.HasChange("user_groups") {
		o, n := d.GetChange("user_groups")
		old := expandStringList(o.(*schema.Set).List())
		newEntries := expandStringList(n.(*schema.Set).List())
		toAdd := difference(newEntries, old)
		toRemove := difference(old, newEntries)

		// Additions
		if len(toAdd) > 0 {
			_, _, err := client.AddUserGroups([]string{tagName}, toAdd)
			if err != nil {
				return err
			}
		}

		// Removals
		if len(toRemove) > 0 {
			_, _, err := client.RemoveUserGroups([]string{tagName}, toRemove)
			if err != nil {
				return err
			}
		}
		d.SetPartial("user_groups")
	}

	if d.HasChange("security_groups") {
		o, n := d.GetChange("securit_groups")
		old := expandStringList(o.(*schema.Set).List())
		newEntries := expandStringList(n.(*schema.Set).List())
		toAdd := difference(newEntries, old)
		toRemove := difference(old, newEntries)

		// Additions
		if len(toAdd) > 0 {
			_, _, err := client.AddSecurityGroups([]string{tagName}, toAdd)
			if err != nil {
				return err
			}
		}

		// Removals
		if len(toRemove) > 0 {
			_, _, err := client.RemoveSecurityGroups([]string{tagName}, toRemove)
			if err != nil {
				return err
			}
		}
		d.SetPartial("security_groups")
	}
	if d.HasChange("protect") {
		protect := d.Get("protect").(bool)
		_, _, err := client.SetProtection(tagName, protect)
		if err != nil {
			return err
		}
		d.SetPartial(("protect"))
	}
	d.Partial(false)
	return nil

}

func resourceContainerHostRead(d *schema.ResourceData, m interface{}) error {
	config := m.(*Config)
	client, err := config.CartelClient()
	if err != nil {
		return err
	}

	tagName := d.Get("name").(string)
	ch, _, err := client.GetDetails(tagName)
	if err != nil {
		return err
	}
	if ch.InstanceID != d.Id() {
		return ErrInstanceIDMismatch
	}
	_ = d.Set("protect", ch.Protection)
	_ = d.Set("volumes", len(ch.BlockDevices))
	_ = d.Set("role", ch.Role)
	_ = d.Set("launch_time", ch.LaunchTime)
	_ = d.Set("block_devices", ch.BlockDevices)
	_ = d.Set("security_groups", ch.SecurityGroups)
	_ = d.Set("user_groups", ch.LdapGroups)
	_ = d.Set("instance_type", ch.InstanceType)
	_ = d.Set("vpc", ch.Vpc)
	_ = d.Set("zone", ch.Zone)
	_ = d.Set("launch_time", ch.LaunchTime)

	return nil
}