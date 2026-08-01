package docker

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
)

// ContainerQuery filters container listings. Zero-value fields do not
// filter; AllEnabled includes stopped containers.
type ContainerQuery struct {
	Labels     Labels
	Name       ContainerName
	AllEnabled bool
}

// containerSummaryWire mirrors one Engine list entry; decode-only.
type containerSummaryWire struct {
	Labels map[string]string
	ID     string `json:"Id"`
	Image  string
	State  string
	Names  []string `json:"Names"`
}

// ContainerSummary is one row of a container listing.
type ContainerSummary struct {
	Labels Labels
	ID     ContainerID
	Name   ContainerName
	Image  ImageRef
	Status ContainerStatus
}

// Containers lists containers matching the query.
func (c Client) Containers(ctx context.Context, query ContainerQuery) ([]ContainerSummary, error) {
	values := url.Values{}
	values.Set("all", strconv.FormatBool(query.AllEnabled))
	if filters, err := listFilters(query.Labels, query.Name); err != nil {
		return nil, err
	} else if filters != "" {
		values.Set("filters", string(filters))
	}
	response, err := c.request(ctx, http.MethodGet, "/containers/json", values, nil, http.StatusOK)
	if err != nil {
		return nil, err
	}
	defer func() { _ = response.Body.Close() }()
	raw, err := readBody(response.Body)
	if err != nil {
		return nil, err
	}
	var wire []containerSummaryWire
	if err := json.Unmarshal(raw, &wire); err != nil {
		return nil, ErrDecodeResponse.With(err)
	}
	return summariesFrom(wire), nil
}

// summariesFrom maps wire list entries onto the public shape.
func summariesFrom(wire []containerSummaryWire) []ContainerSummary {
	summaries := make([]ContainerSummary, len(wire))
	for i, entry := range wire {
		summaries[i] = ContainerSummary{
			ID:     ContainerID(entry.ID),
			Name:   firstName(entry.Names),
			Image:  ImageRef(entry.Image),
			Status: ContainerStatus(entry.State),
			Labels: Labels(entry.Labels),
		}
	}
	return summaries
}

// firstName extracts the primary container name from the Engine's
// slash-prefixed name list.
func firstName(names []string) ContainerName {
	if len(names) == 0 {
		return ""
	}
	return ContainerName(strings.TrimPrefix(names[0], "/"))
}

// filtersDocument is the Engine's JSON filters query parameter.
type filtersDocument string

// listFilters renders the Engine's filters query document from label and
// name terms; empty when there are no terms.
func listFilters(labels Labels, name ContainerName) (filtersDocument, error) {
	terms := map[string][]string{}
	for key, value := range labels {
		terms["label"] = append(terms["label"], string(labelTerm(LabelKey(key), LabelValue(value))))
	}
	slices.Sort(terms["label"])
	if name != "" {
		terms["name"] = []string{string(name)}
	}
	if len(terms) == 0 {
		return "", nil
	}
	encoded, err := marshalBody(terms)
	if err != nil {
		return "", ErrEncodeRequest.With(err)
	}
	return filtersDocument(encoded), nil
}

// LabelKey and LabelValue are the halves of one label entry.
type (
	// LabelKey is a label's key.
	LabelKey string
	// LabelValue is a label's value.
	LabelValue string
	// labelFilterTerm is one rendered label filter term.
	labelFilterTerm string
)

// labelTerm renders one label filter term: bare key when value is empty,
// key=value otherwise.
func labelTerm(key LabelKey, value LabelValue) labelFilterTerm {
	if value == "" {
		return labelFilterTerm(key)
	}
	return labelFilterTerm(string(key) + "=" + string(value))
}
