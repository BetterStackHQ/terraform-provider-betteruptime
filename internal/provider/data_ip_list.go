package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func newIpListDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: ipListLookup,
		Description: "Monitoring IPs lookup.",
		Schema:      ipListSchema,
	}
}

func ipListLookup(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	res, err := meta.(*client).Get(ctx, "/ips-by-cluster.json")
	if err != nil {
		return diag.FromErr(err)
	}
	defer func() {
		// Keep-Alive.
		_, _ = io.Copy(io.Discard, res.Body)
		_ = res.Body.Close()
	}()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return diag.FromErr(err)
	}
	if res.StatusCode != http.StatusOK {
		return diag.Errorf("GET %s returned %d: %s", res.Request.URL.String(), res.StatusCode, string(body))
	}

	// Unmarshal the JSON into a map
	var ipData map[string][]string
	err = json.Unmarshal(body, &ipData)
	if err != nil {
		return diag.FromErr(err)
	}

	// Retrieve the clusters specified by the user from the Terraform resource data
	requestedClusters := d.Get("filter_clusters").([]interface{})
	filterClusters := make(map[string]bool)
	for _, cluster := range requestedClusters {
		filterClusters[cluster.(string)] = true
	}

	// Filter IPs based on the specified clusters, and fetch all clusters
	var filteredIPs []string
	var allClusters []string
	for cluster, ips := range ipData {
		if len(filterClusters) == 0 || filterClusters[cluster] {
			filteredIPs = append(filteredIPs, ips...)
		}
		allClusters = append(allClusters, cluster)
	}

	// Set the data in the Terraform schema
	d.SetId("betterstack_ip_list")
	if err := d.Set("ips", filteredIPs); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("all_clusters", allClusters); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
