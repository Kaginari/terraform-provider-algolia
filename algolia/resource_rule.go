package algolia

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/opt"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mitchellh/mapstructure"
	"strings"
)

func resourceRule() *schema.Resource {

	return &schema.Resource{
		CreateContext: resourceRuleCreate,
		ReadContext:   resourceRuleRead,
		UpdateContext: resourceRuleUpdate,
		DeleteContext: resourceRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"index": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"apply_to_replicas": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},

			"consequence_params": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"condition": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 10,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"anchoring": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validateDiagFunc(validation.StringInSlice([]string{"is", "startsWith", "endsWith", "contains"}, false)),
						},
						"pattern": {
							Type:     schema.TypeString,
							Required: true,
						},
						"alternatives": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
		},
	}
}

func resourceRuleDelete(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	stateID := data.State().ID
	ruleId, indexName, err := resourceRuleParseId(stateID)
	if err != nil {
		return diag.Errorf("unexpected format of ID Error : %s", err)
	}
	index := i.(*apiClient).algolia.InitIndex(indexName)

	forwardToReplicas := opt.ForwardToReplicas(true)

	_, err = index.Exists()

	if err != nil {
		return diag.Errorf("can't find index %s", err)
	}

	_, err = index.DeleteRule(ruleId, forwardToReplicas)

	if err != nil {
		return diag.Errorf("Error deleting index rule : %s ", err)
	}
	return resourceRuleRead(ctx, data, i)
}

func resourceRuleUpdate(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {

	stateID := data.State().ID
	ruleId, indexName, err := resourceRuleParseId(stateID)
	if err != nil {
		return diag.Errorf("unexpected format of ID Error : %s", err)
	}
	index := i.(*apiClient).algolia.InitIndex(indexName)

	_, err = index.Exists()

	if err != nil {
		return diag.Errorf("can't find index %s", err)
	}

	forwardToReplicas := opt.ForwardToReplicas(true)

	_, err = index.DeleteRule(ruleId, forwardToReplicas)

	if err != nil {
		return diag.Errorf("Error deleting index rule : %s ", err)
	}
	indexName = data.Get("index").(string)
	forward := data.Get("apply_to_replicas").(bool)
	ruleId = data.Get("name").(string)
	enabled := data.Get("enabled").(bool)

	forwardToReplicas = opt.ForwardToReplicas(forward)


	consequence := data.Get("consequence_params").(string)

	searchConsequence := search.RuleConsequence{
		Params: &search.RuleParams{
			QueryParams: search.QueryParams{
				Filters: opt.Filters(consequence),
			},
		},
	}
	var ConditionList []RuleCondition
	conditions := data.Get("condition").(*schema.Set).List()
	conditionMapErr := mapstructure.Decode(conditions, &ConditionList)
	if conditionMapErr != nil {
		return diag.Errorf("Error decoding map : %s ", conditionMapErr)
	}
	var listSearchConditions []search.RuleCondition

	if len(ConditionList) != 0 {

		for _, element := range ConditionList {
			var searchCondition search.RuleCondition

			switch element.Anchoring {
			case "is":
				searchCondition.Anchoring = search.Is
			case "startsWith":
				searchCondition.Anchoring = search.StartsWith
			case "endsWith":
				searchCondition.Anchoring = search.EndsWith
			case "contains":
				searchCondition.Anchoring = search.Contains
			}
			searchCondition.Pattern = element.Pattern

			if element.Alternatives {
				searchCondition.Alternatives = search.AlternativesEnabled()
			}

			listSearchConditions = append(listSearchConditions, searchCondition)
		}
	}

	rule := search.Rule{
		ObjectID:   ruleId,
		Conditions: listSearchConditions,
		Consequence: searchConsequence,
		Enabled:    opt.Enabled(enabled),
	}
	_, err = index.SaveRule(rule, forwardToReplicas)

	if err != nil {
		return diag.Errorf("Couldn't create rule: %s ", err)
	}

	return resourceRuleRead(ctx, data, i)
}

func resourceRuleRead(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	stateID := data.State().ID
	ruleId, indexName, err := resourceRuleParseId(stateID)

	if err != nil {
		return diag.Errorf("%s", err)
	}
	index := i.(*apiClient).algolia.InitIndex(indexName)

	_, err = index.Exists()

	if err != nil {
		return diag.Errorf("can't find index %s", err)
	}
	rule, err := index.GetRule(ruleId)

	if err != nil {
		diag.Errorf("can't get rule : %s", err)
	}
	dataSetError := data.Set("enabled", rule.Enabled)
	if dataSetError != nil {
		return diag.Errorf("error setting data : %s ", dataSetError)
	}
	dataSetError = data.Set("apply_to_replicas", data.Get("apply_to_replicas"))
	if dataSetError != nil {
		return diag.Errorf("error setting data : %s ", dataSetError)
	}

	conditions := make([]interface{}, len(rule.Conditions))

	for i, s := range rule.Conditions {
		conditions[i] = map[string]interface{}{
			"pattern":   s.Pattern,
			"anchoring": s.Anchoring,
		}
	}
	dataSetError = data.Set("condition", conditions)
	if dataSetError != nil {
		return diag.Errorf("error setting data : %s ", dataSetError)
	}
	return diags
}

func resourceRuleCreate(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {

	indexName := data.Get("index").(string)
	forward := data.Get("apply_to_replicas").(bool)
	ruleId := data.Get("name").(string)
	enabled := data.Get("enabled").(bool)

	index := i.(*apiClient).algolia.InitIndex(indexName)

	_, err := index.Exists()

	if err != nil {
		return diag.Errorf("can't find index %s", err)
	}

	forwardToReplicas := opt.ForwardToReplicas(forward)

	consequence := data.Get("consequence_params").(string)


	searchConsequence := search.RuleConsequence{
		Params: &search.RuleParams{
			QueryParams: search.QueryParams{
				Filters: opt.Filters(consequence),
			},
		},
	}

	var ConditionList []RuleCondition
	conditions := data.Get("condition").(*schema.Set).List()
	conditionMapErr := mapstructure.Decode(conditions, &ConditionList)
	if conditionMapErr != nil {
		return diag.Errorf("Error decoding map : %s ", conditionMapErr)
	}
	var listSearchConditions []search.RuleCondition

	if len(ConditionList) != 0 {

		for _, element := range ConditionList {
			var searchCondition search.RuleCondition

			switch element.Anchoring {
			case "is":
				searchCondition.Anchoring = search.Is
			case "startsWith":
				searchCondition.Anchoring = search.StartsWith
			case "endsWith":
				searchCondition.Anchoring = search.EndsWith
			case "contains":
				searchCondition.Anchoring = search.Contains
			}
			searchCondition.Pattern = element.Pattern
			if element.Alternatives {
				searchCondition.Alternatives = search.AlternativesEnabled()
			}

			listSearchConditions = append(listSearchConditions, searchCondition)
		}
	}

	rule := search.Rule{
		ObjectID:    ruleId,
		Conditions:  listSearchConditions,
		Consequence: searchConsequence,
		Enabled:     opt.Enabled(enabled),
	}
	_, err = index.SaveRule(rule, forwardToReplicas)

	if err != nil {
		return diag.Errorf("Couldn't create rule: %s ", err)
	}

	str := indexName + "." + ruleId
	encoded := base64.StdEncoding.EncodeToString([]byte(str))
	data.SetId(encoded)

	return resourceRuleRead(ctx, data, i)
}

func validateDiagFunc(validateFunc func(interface{}, string) ([]string, []error)) schema.SchemaValidateDiagFunc {
	return func(i interface{}, path cty.Path) diag.Diagnostics {
		warnings, errs := validateFunc(i, fmt.Sprintf("%+v", path))
		var diags diag.Diagnostics
		for _, warning := range warnings {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  warning,
			})
		}
		for _, err := range errs {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  err.Error(),
			})
		}
		return diags
	}
}

func resourceRuleParseId(id string) (string, string, error) {
	result, errEncoding := base64.StdEncoding.DecodeString(id)

	if errEncoding != nil {
		return "", "", fmt.Errorf("unexpected format of ID Error : %s", errEncoding)
	}
	parts := strings.SplitN(string(result), ".", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected attribute1.attribute2", id)
	}

	index := parts[0]
	ruleId := parts[1]

	return index, ruleId, nil
}

type RuleCondition struct {
	Anchoring    string `json:"anchoring"`
	Pattern      string `json:"pattern"`
	Alternatives bool   `json:"alternatives"`
}
