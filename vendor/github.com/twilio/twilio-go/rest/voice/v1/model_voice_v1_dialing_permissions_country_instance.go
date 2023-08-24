/*
 * This code was generated by
 * ___ _ _ _ _ _    _ ____    ____ ____ _    ____ ____ _  _ ____ ____ ____ ___ __   __
 *  |  | | | | |    | |  | __ |  | |__| | __ | __ |___ |\ | |___ |__/ |__|  | |  | |__/
 *  |  |_|_| | |___ | |__|    |__| |  | |    |__] |___ | \| |___ |  \ |  |  | |__| |  \
 *
 * Twilio - Voice
 * This is the public Twilio REST API.
 *
 * NOTE: This class is auto generated by OpenAPI Generator.
 * https://openapi-generator.tech
 * Do not edit the class manually.
 */

package openapi

// VoiceV1DialingPermissionsCountryInstance struct for VoiceV1DialingPermissionsCountryInstance
type VoiceV1DialingPermissionsCountryInstance struct {
	// The [ISO country code](https://en.wikipedia.org/wiki/ISO_3166-1_alpha-2).
	IsoCode *string `json:"iso_code,omitempty"`
	// The name of the country.
	Name *string `json:"name,omitempty"`
	// The name of the continent in which the country is located.
	Continent *string `json:"continent,omitempty"`
	// The E.164 assigned [country codes(s)](https://www.itu.int/itudoc/itu-t/ob-lists/icc/e164_763.html)
	CountryCodes *[]string `json:"country_codes,omitempty"`
	// Whether dialing to low-risk numbers is enabled.
	LowRiskNumbersEnabled *bool `json:"low_risk_numbers_enabled,omitempty"`
	// Whether dialing to high-risk special services numbers is enabled. These prefixes include number ranges allocated by the country and include premium numbers, special services, shared cost, and others
	HighRiskSpecialNumbersEnabled *bool `json:"high_risk_special_numbers_enabled,omitempty"`
	// Whether dialing to high-risk [toll fraud](https://www.twilio.com/learn/voice-and-video/toll-fraud) numbers is enabled. These prefixes include narrow number ranges that have a high-risk of international revenue sharing fraud (IRSF) attacks, also known as [toll fraud](https://www.twilio.com/learn/voice-and-video/toll-fraud). These prefixes are collected from anti-fraud databases and verified by analyzing calls on our network. These prefixes are not available for download and are updated frequently
	HighRiskTollfraudNumbersEnabled *bool `json:"high_risk_tollfraud_numbers_enabled,omitempty"`
	// The absolute URL of this resource.
	Url *string `json:"url,omitempty"`
	// A list of URLs related to this resource.
	Links *map[string]interface{} `json:"links,omitempty"`
}
