// Package forms provides the forms resource client for the 3Common API.
//
// Most callers reach this package through
// [github.com/3-Common/sdk/sdk-go/client.API.Forms]:
//
//	api, _ := client.New(threecommon.Config{APIKey: "..."})
//	page, err := api.Forms.List(ctx, &forms.ListParams{
//		Type:     forms.TypeStandalone,
//		PageSize: threecommon.Int(50),
//	})
//
// To use the resource standalone (without the aggregator), call [New]:
//
//	cli, _ := forms.New(threecommon.Config{APIKey: "..."})
//	page, err := cli.List(ctx, nil)
//
// A form is a tree of elements (questions, static content) arranged into rows.
// The element-level methods (AddElement, UpdateElement, MoveElement, the
// logic-rule and "other"-option toggles) mutate that tree. Type names inside
// this package omit the "Form" prefix to avoid stutter (e.g. forms.ListParams,
// not forms.FormListParams); the [Form] document and its compact [FormSummary]
// list projection are the exception, named to match the other SDKs.
package forms

import (
	threecommon "github.com/3-Common/sdk/sdk-go"
)

// Type is the kind of form. Standalone forms are independent; order forms are
// attached to a checkout flow.
type Type string

// Type values.
const (
	TypeStandalone Type = "standalone"
	TypeOrder      Type = "order"
)

// Status is the lifecycle status of a form.
type Status string

// Status values.
const (
	StatusDraft    Status = "draft"
	StatusActive   Status = "active"
	StatusArchived Status = "archived"
)

// SubmitButtonWidth controls how the submit button sizes itself.
type SubmitButtonWidth string

// SubmitButtonWidth values.
const (
	SubmitButtonWidthAuto SubmitButtonWidth = "auto"
	SubmitButtonWidthFill SubmitButtonWidth = "fill"
)

// SubmitButtonAlign controls the submit button's horizontal alignment.
type SubmitButtonAlign string

// SubmitButtonAlign values.
const (
	SubmitButtonAlignLeft   SubmitButtonAlign = "left"
	SubmitButtonAlignCenter SubmitButtonAlign = "center"
)

// ElementType is the kind of a form element. Unknown values from a future API
// version surface as the raw string.
type ElementType string

// ElementType values.
const (
	ElementText                ElementType = "Text"
	ElementMultiLineText       ElementType = "Multi-line Text"
	ElementSelectOne           ElementType = "Select One"
	ElementSelectOneOther      ElementType = `Select One or "Other"`
	ElementSelectMultiple      ElementType = "Select Multiple"
	ElementSelectMultipleOther ElementType = `Select Multiple with "Other"`
	ElementYesNo               ElementType = "Yes/No"
	ElementDate                ElementType = "Date"
	ElementFile                ElementType = "File"
	ElementEmail               ElementType = "Email"
	ElementPhone               ElementType = "Phone"
	ElementStaticText          ElementType = "Static Text"
	ElementStaticImage         ElementType = "Static Image"
)

// LogicOperator is the operator a selection-question logic rule applies to its
// option indices.
type LogicOperator string

// LogicOperator values.
const (
	LogicOperatorAllOf  LogicOperator = "all_of"
	LogicOperatorAnyOf  LogicOperator = "any_of"
	LogicOperatorNoneOf LogicOperator = "none_of"
)

// SelectionType is the comparison applied by a Yes/No logic rule.
type SelectionType string

// SelectionType values.
const (
	SelectionTypeIs    SelectionType = "is"
	SelectionTypeIsNot SelectionType = "is_not"
)

// MoveSection is the section an element is moved into on an order form.
type MoveSection string

// MoveSection values.
const (
	MoveSectionBuyer        MoveSection = "buyer"
	MoveSectionTicketHolder MoveSection = "ticket-holder"
)

// FormSummary is the compact projection returned by [Client.List] in each
// page's data array.
type FormSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	NumElements int    `json:"numElements"`
	Type        Type   `json:"type"`
	Status      Status `json:"status"`
}

// Form is the full form projection returned by Retrieve, Create, Update,
// Duplicate, and MoveElement. It carries the form's settings plus its element
// tree.
type Form struct {
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	NameHidden        bool              `json:"nameHidden,omitempty"`
	OwnerID           string            `json:"ownerId"`
	Status            Status            `json:"status"`
	Type              Type              `json:"type"`
	SubmitButtonText  string            `json:"submitButtonText,omitempty"`
	SubmitButtonWidth SubmitButtonWidth `json:"submitButtonWidth,omitempty"`
	SubmitButtonAlign SubmitButtonAlign `json:"submitButtonAlign,omitempty"`
	// Rows describes the form's layout grid. Its shape is loosely typed
	// because layout details are not part of the stable SDK surface.
	Rows     []map[string]any `json:"rows,omitempty"`
	Elements []Element        `json:"elements,omitempty"`
	// AttendeeRowsStart is only present on order forms; it marks where the
	// per-attendee section begins.
	AttendeeRowsStart *int `json:"attendeeRowsStart,omitempty"`
}

// LogicGroup is one conditional-visibility rule attached to a selection or
// Yes/No element. Selection rules use OptionIndices + Operator; Yes/No rules
// use SelectionType + Value.
type LogicGroup struct {
	RevealedElementIndex *int          `json:"revealedElementIndex,omitempty"`
	OptionIndices        []int         `json:"optionIndices,omitempty"`
	Operator             LogicOperator `json:"operator,omitempty"`
	SelectionType        SelectionType `json:"selectionType,omitempty"`
	Value                *bool         `json:"value,omitempty"`
}

// EventItemRefType discriminates the variants of [EventItemRef].
type EventItemRefType string

// EventItemRefType values.
const (
	EventItemRefEventItem       EventItemRefType = "eventItem"
	EventItemRefEventProduct    EventItemRefType = "eventProduct"
	EventItemRefCheckoutProduct EventItemRefType = "checkoutProduct"
)

// EventItemRef points at one purchasable item that gates an order-form
// element's visibility. Type selects the variant and which ID fields apply:
// "eventItem" uses EventID + ItemID, "eventProduct" uses EventID + ProductID,
// and "checkoutProduct" uses CheckoutID + ProductID.
type EventItemRef struct {
	Type       EventItemRefType `json:"type"`
	EventID    string           `json:"eventId,omitempty"`
	ItemID     string           `json:"itemId,omitempty"`
	ProductID  string           `json:"productId,omitempty"`
	CheckoutID string           `json:"checkoutId,omitempty"`
}

// Element is a single form element. It is a wide union over every element type
// (text, selection, date, file, static content, ...); only the fields relevant
// to a given Type are populated.
type Element struct {
	ID           string         `json:"id"`
	Type         ElementType    `json:"type"`
	Prompt       string         `json:"prompt,omitempty"`
	PromptHidden bool           `json:"promptHidden,omitempty"`
	HelperText   string         `json:"helperText,omitempty"`
	Placeholder  string         `json:"placeholder,omitempty"`
	Required     *bool          `json:"required,omitempty"`
	PropertyID   string         `json:"propertyId,omitempty"`
	PropertyData map[string]any `json:"propertyData,omitempty"`
	ContactField string         `json:"contactField,omitempty"`
	Options      []string       `json:"options,omitempty"`
	Dropdown     *bool          `json:"dropdown,omitempty"`
	OtherPrompt  string         `json:"otherPrompt,omitempty"`
	MinChoices   *int           `json:"minChoices,omitempty"`
	MaxChoices   *int           `json:"maxChoices,omitempty"`
	// Min and Max are the earliest/latest selectable date for Date elements,
	// in YYYY-MM-DD format.
	Min         *string      `json:"min,omitempty"`
	Max         *string      `json:"max,omitempty"`
	Accept      string       `json:"accept,omitempty"`
	LogicGroups []LogicGroup `json:"logicGroups,omitempty"`
	Content     string       `json:"content,omitempty"`
	ImageURL    string       `json:"imageUrl,omitempty"`
	ImageWidth  *int         `json:"imageWidth,omitempty"`
	// ForEventItems restricts an order-form element to appear only when at
	// least one of the referenced items is in the cart.
	ForEventItems []EventItemRef `json:"forEventItems,omitempty"`
	// AskAllAttendees is only present on order forms. True means the element
	// sits in the per-attendee section and repeats for every ticket.
	AskAllAttendees *bool `json:"askAllAttendees,omitempty"`
}

// ListParams are the query parameters accepted by [Client.List] and
// [Client.ListAutoPaginate].
type ListParams struct {
	// Page is the 0-indexed page number. Nil uses the server default (0).
	Page *int

	// PageSize is the items-per-page cap. Nil uses the server default (50);
	// the server caps at 250.
	PageSize *int

	// Type restricts results to a single form type. Empty fetches all types.
	Type Type
}

// CreateParams is the body shape accepted by [Client.Create]. Name and Type
// are required.
type CreateParams struct {
	Name              string            `json:"name"`
	Type              Type              `json:"type"`
	NameHidden        *bool             `json:"nameHidden,omitempty"`
	Status            Status            `json:"status,omitempty"`
	SubmitButtonText  string            `json:"submitButtonText,omitempty"`
	SubmitButtonWidth SubmitButtonWidth `json:"submitButtonWidth,omitempty"`
	SubmitButtonAlign SubmitButtonAlign `json:"submitButtonAlign,omitempty"`
}

// UpdateParams is the body shape accepted by [Client.Update]. Every field is
// optional: nil fields are omitted from the request and only the fields you
// set are changed. Pointer strings distinguish "unset" from an explicit empty
// string. NameHidden and SubmitButtonAlign accept [threecommon.Null] to send
// an explicit JSON null, which clears the setting server-side.
type UpdateParams struct {
	Name              *string                                  `json:"name,omitempty"`
	NameHidden        *threecommon.Nullable[bool]              `json:"nameHidden,omitempty"`
	Status            Status                                   `json:"status,omitempty"`
	SubmitButtonText  *string                                  `json:"submitButtonText,omitempty"`
	SubmitButtonWidth SubmitButtonWidth                        `json:"submitButtonWidth,omitempty"`
	SubmitButtonAlign *threecommon.Nullable[SubmitButtonAlign] `json:"submitButtonAlign,omitempty"`
}

// DuplicateParams is the body shape accepted by [Client.Duplicate]. Both
// fields are optional overrides for the copy.
type DuplicateParams struct {
	Name   string `json:"name,omitempty"`
	Status Status `json:"status,omitempty"`
}

// AddElementParams is the body shape accepted by [Client.AddElement]. It is a
// wide union over every element type; set only the fields relevant to Type.
type AddElementParams struct {
	Type         ElementType    `json:"type"`
	Prompt       string         `json:"prompt,omitempty"`
	PromptHidden *bool          `json:"promptHidden,omitempty"`
	HelperText   string         `json:"helperText,omitempty"`
	Placeholder  string         `json:"placeholder,omitempty"`
	Required     *bool          `json:"required,omitempty"`
	PropertyID   string         `json:"propertyId,omitempty"`
	PropertyData map[string]any `json:"propertyData,omitempty"`
	ContactField string         `json:"contactField,omitempty"`
	Options      []string       `json:"options,omitempty"`
	Dropdown     *bool          `json:"dropdown,omitempty"`
	OtherPrompt  string         `json:"otherPrompt,omitempty"`
	MinChoices   *int           `json:"minChoices,omitempty"`
	MaxChoices   *int           `json:"maxChoices,omitempty"`
	// Min and Max are the earliest/latest selectable date for Date elements,
	// in YYYY-MM-DD format.
	Min         *string      `json:"min,omitempty"`
	Max         *string      `json:"max,omitempty"`
	Accept      string       `json:"accept,omitempty"`
	LogicGroups []LogicGroup `json:"logicGroups,omitempty"`
	Content     string       `json:"content,omitempty"`
	ImageURL    string       `json:"imageUrl,omitempty"`
	ImageWidth  *int         `json:"imageWidth,omitempty"`
}

// UpdateElementParams is the body shape accepted by [Client.UpdateElement].
// Every field is optional: nil fields are omitted from the request and only
// the fields you set are changed. Every field except Prompt is nullable
// server-side: pass [threecommon.Null] to send an explicit JSON null, which
// clears the setting; use [threecommon.NullableOf] to set a concrete value.
type UpdateElementParams struct {
	Prompt       *string                         `json:"prompt,omitempty"`
	PromptHidden *threecommon.Nullable[bool]     `json:"promptHidden,omitempty"`
	HelperText   *threecommon.Nullable[string]   `json:"helperText,omitempty"`
	Placeholder  *threecommon.Nullable[string]   `json:"placeholder,omitempty"`
	Required     *threecommon.Nullable[bool]     `json:"required,omitempty"`
	PropertyID   *threecommon.Nullable[string]   `json:"propertyId,omitempty"`
	ContactField *threecommon.Nullable[string]   `json:"contactField,omitempty"`
	Options      *threecommon.Nullable[[]string] `json:"options,omitempty"`
	Dropdown     *threecommon.Nullable[bool]     `json:"dropdown,omitempty"`
	OtherPrompt  *threecommon.Nullable[string]   `json:"otherPrompt,omitempty"`
	MinChoices   *threecommon.Nullable[int]      `json:"minChoices,omitempty"`
	MaxChoices   *threecommon.Nullable[int]      `json:"maxChoices,omitempty"`
	// Min and Max are the earliest/latest selectable date for Date elements,
	// in YYYY-MM-DD format.
	Min           *threecommon.Nullable[string]         `json:"min,omitempty"`
	Max           *threecommon.Nullable[string]         `json:"max,omitempty"`
	Accept        *threecommon.Nullable[string]         `json:"accept,omitempty"`
	LogicGroups   *threecommon.Nullable[[]LogicGroup]   `json:"logicGroups,omitempty"`
	Content       *threecommon.Nullable[string]         `json:"content,omitempty"`
	ImageURL      *threecommon.Nullable[string]         `json:"imageUrl,omitempty"`
	ImageWidth    *threecommon.Nullable[int]            `json:"imageWidth,omitempty"`
	ForEventItems *threecommon.Nullable[[]EventItemRef] `json:"forEventItems,omitempty"`
}

// MoveElementParams is the body shape accepted by [Client.MoveElement].
// Position is required.
type MoveElementParams struct {
	Position int         `json:"position"`
	Section  MoveSection `json:"section,omitempty"`
}

// LogicCondition is the trigger for a logic rule added via
// [Client.AddLogicRule]. For selection questions set OptionIndices + Operator;
// for Yes/No questions set SelectionType + Value.
type LogicCondition struct {
	OptionIndices []int         `json:"optionIndices,omitempty"`
	Operator      LogicOperator `json:"operator,omitempty"`
	SelectionType SelectionType `json:"selectionType,omitempty"`
	Value         *bool         `json:"value,omitempty"`
}

// AddLogicRuleParams is the body shape accepted by [Client.AddLogicRule]. Both
// fields are required.
type AddLogicRuleParams struct {
	RevealedElementID string         `json:"revealedElementId"`
	Condition         LogicCondition `json:"condition"`
}

// EnableOtherOptionParams is the body shape accepted by
// [Client.EnableOtherOption]. OtherPrompt is required.
type EnableOtherOptionParams struct {
	OtherPrompt string `json:"otherPrompt"`
}

// ListResponse is the body returned by GET /v1/forms.
type ListResponse struct {
	Data    []FormSummary `json:"data"`
	HasMore bool          `json:"hasMore"`
}

// DeleteElementResult is the data shape unwrapped from
// DELETE /v1/forms/{id}/elements/{elementId}. Echoes the removed element's id.
type DeleteElementResult struct {
	DeletedElementID string `json:"deletedElementId"`
}

// formEnvelope is the {"data": Form} shape used by form-returning
// endpoints.
type formEnvelope struct {
	Data Form `json:"data"`
}

// elementEnvelope is the {"data": Element} shape used by element-returning
// endpoints.
type elementEnvelope struct {
	Data Element `json:"data"`
}

// deleteElementEnvelope is the {"data": {deletedElementId}} shape used by
// DeleteElement.
type deleteElementEnvelope struct {
	Data DeleteElementResult `json:"data"`
}
