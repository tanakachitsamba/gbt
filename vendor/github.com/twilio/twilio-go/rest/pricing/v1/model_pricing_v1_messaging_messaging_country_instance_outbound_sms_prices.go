/*
 * This code was generated by
 * ___ _ _ _ _ _    _ ____    ____ ____ _    ____ ____ _  _ ____ ____ ____ ___ __   __
 *  |  | | | | |    | |  | __ |  | |__| | __ | __ |___ |\ | |___ |__/ |__|  | |  | |__/
 *  |  |_|_| | |___ | |__|    |__| |  | |    |__] |___ | \| |___ |  \ |  |  | |__| |  \
 *
 * Twilio - Pricing
 * This is the public Twilio REST API.
 *
 * NOTE: This class is auto generated by OpenAPI Generator.
 * https://openapi-generator.tech
 * Do not edit the class manually.
 */

package openapi

// PricingV1MessagingMessagingCountryInstanceOutboundSmsPrices struct for PricingV1MessagingMessagingCountryInstanceOutboundSmsPrices
type PricingV1MessagingMessagingCountryInstanceOutboundSmsPrices struct {
	Carrier string                                                              `json:"carrier,omitempty"`
	Mcc     string                                                              `json:"mcc,omitempty"`
	Mnc     string                                                              `json:"mnc,omitempty"`
	Prices  []PricingV1MessagingMessagingCountryInstanceOutboundSmsPricesPrices `json:"prices,omitempty"`
}