// Package railway provides a client for Railway's GraphQL API
package railway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const apiEndpoint = "https://backboard.railway.com/graphql/v2"

// Client for Railway API
type Client struct {
	token     string
	projectID string
	serviceID string
	envID     string
	http      *http.Client
}

// New creates a new Railway API client
func New(token, projectID, serviceID, envID string) *Client {
	return &Client{
		token:     token,
		projectID: projectID,
		serviceID: serviceID,
		envID:     envID,
		http:      &http.Client{},
	}
}

// graphQLRequest represents a GraphQL request
type graphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

// graphQLResponse represents a GraphQL response
type graphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

// doRequest executes a GraphQL request
func (c *Client) doRequest(ctx context.Context, query string, variables map[string]any) (json.RawMessage, error) {
	reqBody := graphQLRequest{
		Query:     query,
		Variables: variables,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiEndpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var gqlResp graphQLResponse
	if err := json.Unmarshal(respBody, &gqlResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL error: %s", gqlResp.Errors[0].Message)
	}

	return gqlResp.Data, nil
}

// GetLatestDeployment fetches the latest active deployment ID
func (c *Client) GetLatestDeployment(ctx context.Context) (string, error) {
	query := `
		query deployments($projectId: String!, $environmentId: String!, $serviceId: String!) {
			deployments(
				first: 1
				input: {
					projectId: $projectId
					environmentId: $environmentId
					serviceId: $serviceId
				}
			) {
				edges {
					node {
						id
					}
				}
			}
		}
	`

	variables := map[string]any{
		"projectId":     c.projectID,
		"environmentId": c.envID,
		"serviceId":     c.serviceID,
	}

	data, err := c.doRequest(ctx, query, variables)
	if err != nil {
		return "", err
	}

	var result struct {
		Deployments struct {
			Edges []struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"deployments"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse deployments: %w", err)
	}

	if len(result.Deployments.Edges) == 0 {
		return "", fmt.Errorf("no deployments found")
	}

	return result.Deployments.Edges[0].Node.ID, nil
}

// RestartDeployment triggers a restart of the specified deployment
func (c *Client) RestartDeployment(ctx context.Context, deploymentID string) error {
	query := `
		mutation deploymentRestart($id: String!) {
			deploymentRestart(id: $id)
		}
	`

	variables := map[string]any{
		"id": deploymentID,
	}

	_, err := c.doRequest(ctx, query, variables)
	return err
}
