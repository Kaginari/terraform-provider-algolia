package algolia

import (
	"context"
	"errors"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strconv"
	"time"
)


func resourceIndex() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIndexCreate,
		ReadContext:   resourceIndexRead,
		UpdateContext: resourceIndexUpdate,
		DeleteContext: resourceIndexDelete,
		Importer: &schema.ResourceImporter{
			StateContext: importIndexState,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceIndexDelete(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	index := i.(*search.Client).InitIndex(getAlgoliaIndex(data))
	res,err := index.Exists()
	if err != nil {
		return diag.FromErr(err)
	}
	if res == false{
		diag.FromErr(errors.New("Index doesn't Exists "))
	}
	index.Delete()
	return diags
}

func resourceIndexRead(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {

	var diags diag.Diagnostics
	//if err:=
	diags = nil
	return diags
}

func resourceIndexCreate(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	index := i.(*search.Client).InitIndex(getAlgoliaIndex(data))
	res,err := index.Exists()
	if err != nil {
		return diag.FromErr(err)
	}
	if res == true{
		diag.FromErr(errors.New("Index Exists "))
	}

	data.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	index.SaveObject(search.ObjectIterator{})
	return nil
}
func resourceIndexUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	index := m.(*search.Client).InitIndex(getAlgoliaIndex(d))
	res,err := index.Exists()
	if err != nil {
		return diag.FromErr(err)
	}
	if res == false{
		diag.FromErr(errors.New("Index doesn't Exists "))
	}
	index.SetSettings(index.GetSettings())

	return resourceIndexRead(ctx, d, m)
}

func getAlgoliaIndex(data *schema.ResourceData) string {
	return data.Get("name").(string)
}

func importIndexState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	if err := d.Set("name", d.Get("name")); err != nil {
		return nil, err
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return []*schema.ResourceData{d}, nil
}
