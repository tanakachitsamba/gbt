/*
 * This code was generated by
 * ___ _ _ _ _ _    _ ____    ____ ____ _    ____ ____ _  _ ____ ____ ____ ___ __   __
 *  |  | | | | |    | |  | __ |  | |__| | __ | __ |___ |\ | |___ |__/ |__|  | |  | |__/
 *  |  |_|_| | |___ | |__|    |__| |  | |    |__] |___ | \| |___ |  \ |  |  | |__| |  \
 *
 * Twilio - Verify
 * This is the public Twilio REST API.
 *
 * NOTE: This class is auto generated by OpenAPI Generator.
 * https://openapi-generator.tech
 * Do not edit the class manually.
 */

package openapi

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/twilio/twilio-go/client"
)

// Optional parameters for the method 'CreateMessagingConfiguration'
type CreateMessagingConfigurationParams struct {
	// The [ISO-3166-1](https://en.wikipedia.org/wiki/ISO_3166-1_alpha-2) country code of the country this configuration will be applied to. If this is a global configuration, Country will take the value `all`.
	Country *string `json:"Country,omitempty"`
	// The SID of the [Messaging Service](https://www.twilio.com/docs/messaging/services/api) to be used to send SMS to the country of this configuration.
	MessagingServiceSid *string `json:"MessagingServiceSid,omitempty"`
}

func (params *CreateMessagingConfigurationParams) SetCountry(Country string) *CreateMessagingConfigurationParams {
	params.Country = &Country
	return params
}
func (params *CreateMessagingConfigurationParams) SetMessagingServiceSid(MessagingServiceSid string) *CreateMessagingConfigurationParams {
	params.MessagingServiceSid = &MessagingServiceSid
	return params
}

// Create a new MessagingConfiguration for a service.
func (c *ApiService) CreateMessagingConfiguration(ServiceSid string, params *CreateMessagingConfigurationParams) (*VerifyV2MessagingConfiguration, error) {
	path := "/v2/Services/{ServiceSid}/MessagingConfigurations"
	path = strings.Replace(path, "{"+"ServiceSid"+"}", ServiceSid, -1)

	data := url.Values{}
	headers := make(map[string]interface{})

	if params != nil && params.Country != nil {
		data.Set("Country", *params.Country)
	}
	if params != nil && params.MessagingServiceSid != nil {
		data.Set("MessagingServiceSid", *params.MessagingServiceSid)
	}

	resp, err := c.requestHandler.Post(c.baseURL+path, data, headers)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	ps := &VerifyV2MessagingConfiguration{}
	if err := json.NewDecoder(resp.Body).Decode(ps); err != nil {
		return nil, err
	}

	return ps, err
}

// Delete a specific MessagingConfiguration.
func (c *ApiService) DeleteMessagingConfiguration(ServiceSid string, Country string) error {
	path := "/v2/Services/{ServiceSid}/MessagingConfigurations/{Country}"
	path = strings.Replace(path, "{"+"ServiceSid"+"}", ServiceSid, -1)
	path = strings.Replace(path, "{"+"Country"+"}", Country, -1)

	data := url.Values{}
	headers := make(map[string]interface{})

	resp, err := c.requestHandler.Delete(c.baseURL+path, data, headers)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}

// Fetch a specific MessagingConfiguration.
func (c *ApiService) FetchMessagingConfiguration(ServiceSid string, Country string) (*VerifyV2MessagingConfiguration, error) {
	path := "/v2/Services/{ServiceSid}/MessagingConfigurations/{Country}"
	path = strings.Replace(path, "{"+"ServiceSid"+"}", ServiceSid, -1)
	path = strings.Replace(path, "{"+"Country"+"}", Country, -1)

	data := url.Values{}
	headers := make(map[string]interface{})

	resp, err := c.requestHandler.Get(c.baseURL+path, data, headers)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	ps := &VerifyV2MessagingConfiguration{}
	if err := json.NewDecoder(resp.Body).Decode(ps); err != nil {
		return nil, err
	}

	return ps, err
}

// Optional parameters for the method 'ListMessagingConfiguration'
type ListMessagingConfigurationParams struct {
	// How many resources to return in each list page. The default is 50, and the maximum is 1000.
	PageSize *int `json:"PageSize,omitempty"`
	// Max number of records to return.
	Limit *int `json:"limit,omitempty"`
}

func (params *ListMessagingConfigurationParams) SetPageSize(PageSize int) *ListMessagingConfigurationParams {
	params.PageSize = &PageSize
	return params
}
func (params *ListMessagingConfigurationParams) SetLimit(Limit int) *ListMessagingConfigurationParams {
	params.Limit = &Limit
	return params
}

// Retrieve a single page of MessagingConfiguration records from the API. Request is executed immediately.
func (c *ApiService) PageMessagingConfiguration(ServiceSid string, params *ListMessagingConfigurationParams, pageToken, pageNumber string) (*ListMessagingConfigurationResponse, error) {
	path := "/v2/Services/{ServiceSid}/MessagingConfigurations"

	path = strings.Replace(path, "{"+"ServiceSid"+"}", ServiceSid, -1)

	data := url.Values{}
	headers := make(map[string]interface{})

	if params != nil && params.PageSize != nil {
		data.Set("PageSize", fmt.Sprint(*params.PageSize))
	}

	if pageToken != "" {
		data.Set("PageToken", pageToken)
	}
	if pageNumber != "" {
		data.Set("Page", pageNumber)
	}

	resp, err := c.requestHandler.Get(c.baseURL+path, data, headers)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	ps := &ListMessagingConfigurationResponse{}
	if err := json.NewDecoder(resp.Body).Decode(ps); err != nil {
		return nil, err
	}

	return ps, err
}

// Lists MessagingConfiguration records from the API as a list. Unlike stream, this operation is eager and loads 'limit' records into memory before returning.
func (c *ApiService) ListMessagingConfiguration(ServiceSid string, params *ListMessagingConfigurationParams) ([]VerifyV2MessagingConfiguration, error) {
	response, errors := c.StreamMessagingConfiguration(ServiceSid, params)

	records := make([]VerifyV2MessagingConfiguration, 0)
	for record := range response {
		records = append(records, record)
	}

	if err := <-errors; err != nil {
		return nil, err
	}

	return records, nil
}

// Streams MessagingConfiguration records from the API as a channel stream. This operation lazily loads records as efficiently as possible until the limit is reached.
func (c *ApiService) StreamMessagingConfiguration(ServiceSid string, params *ListMessagingConfigurationParams) (chan VerifyV2MessagingConfiguration, chan error) {
	if params == nil {
		params = &ListMessagingConfigurationParams{}
	}
	params.SetPageSize(client.ReadLimits(params.PageSize, params.Limit))

	recordChannel := make(chan VerifyV2MessagingConfiguration, 1)
	errorChannel := make(chan error, 1)

	response, err := c.PageMessagingConfiguration(ServiceSid, params, "", "")
	if err != nil {
		errorChannel <- err
		close(recordChannel)
		close(errorChannel)
	} else {
		go c.streamMessagingConfiguration(response, params, recordChannel, errorChannel)
	}

	return recordChannel, errorChannel
}

func (c *ApiService) streamMessagingConfiguration(response *ListMessagingConfigurationResponse, params *ListMessagingConfigurationParams, recordChannel chan VerifyV2MessagingConfiguration, errorChannel chan error) {
	curRecord := 1

	for response != nil {
		responseRecords := response.MessagingConfigurations
		for item := range responseRecords {
			recordChannel <- responseRecords[item]
			curRecord += 1
			if params.Limit != nil && *params.Limit < curRecord {
				close(recordChannel)
				close(errorChannel)
				return
			}
		}

		record, err := client.GetNext(c.baseURL, response, c.getNextListMessagingConfigurationResponse)
		if err != nil {
			errorChannel <- err
			break
		} else if record == nil {
			break
		}

		response = record.(*ListMessagingConfigurationResponse)
	}

	close(recordChannel)
	close(errorChannel)
}

func (c *ApiService) getNextListMessagingConfigurationResponse(nextPageUrl string) (interface{}, error) {
	if nextPageUrl == "" {
		return nil, nil
	}
	resp, err := c.requestHandler.Get(nextPageUrl, nil, nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	ps := &ListMessagingConfigurationResponse{}
	if err := json.NewDecoder(resp.Body).Decode(ps); err != nil {
		return nil, err
	}
	return ps, nil
}

// Optional parameters for the method 'UpdateMessagingConfiguration'
type UpdateMessagingConfigurationParams struct {
	// The SID of the [Messaging Service](https://www.twilio.com/docs/messaging/services/api) to be used to send SMS to the country of this configuration.
	MessagingServiceSid *string `json:"MessagingServiceSid,omitempty"`
}

func (params *UpdateMessagingConfigurationParams) SetMessagingServiceSid(MessagingServiceSid string) *UpdateMessagingConfigurationParams {
	params.MessagingServiceSid = &MessagingServiceSid
	return params
}

// Update a specific MessagingConfiguration
func (c *ApiService) UpdateMessagingConfiguration(ServiceSid string, Country string, params *UpdateMessagingConfigurationParams) (*VerifyV2MessagingConfiguration, error) {
	path := "/v2/Services/{ServiceSid}/MessagingConfigurations/{Country}"
	path = strings.Replace(path, "{"+"ServiceSid"+"}", ServiceSid, -1)
	path = strings.Replace(path, "{"+"Country"+"}", Country, -1)

	data := url.Values{}
	headers := make(map[string]interface{})

	if params != nil && params.MessagingServiceSid != nil {
		data.Set("MessagingServiceSid", *params.MessagingServiceSid)
	}

	resp, err := c.requestHandler.Post(c.baseURL+path, data, headers)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	ps := &VerifyV2MessagingConfiguration{}
	if err := json.NewDecoder(resp.Body).Decode(ps); err != nil {
		return nil, err
	}

	return ps, err
}