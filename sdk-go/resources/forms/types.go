// Package forms provides the forms resource client for the 3Common API.
//
// Most callers reach this package through
// [github.com/3-Common/sdk/sdk-go/client.API.Forms]:
//
//	api, _ := client.New(threecommon.Config{APIKey: "..."})
//	form, err := api.Forms.Create(ctx, &forms.CreateParams{
//		Name: "Customer survey",
//		Type: forms.FormTypeStandalone,
//	})
//
// To use the resource standalone (without the aggregator), call [New]:
//
//	cli, _ := forms.New(threecommon.Config{APIKey: "..."})
//	page, err := cli.List(ctx, nil)
//
// A form is a tree of elements (questions) laid out across rows. The list
// endpoint returns the compact [FormSummary] projection; every detail and
// mutation endpoint returns the full [Form]. Element-level mutations return
// the affected [Element].
//
// Type names inside this package omit the "Form" prefix to avoid stutter
// (e.g. forms.ListParams, not forms.FormListParams).
package forms

// FormType distinguishes a standalone form from one attached to an event
// order flow.
type FormType string

// FormType values.
const (
	// FormTypeStandalone is a form accessed directly at its own URL.
	FormTypeStandalone FormType = "standalone"
	// FormTypeOrder is a form embedded in an event's checkout/order flow.
	FormTypeOrder FormType = "order"
)

// FormStatus is the lifecycle status of a form.
//
//   - FormStatusDraft: not reachable at its URL, hidden from default tables
//   - FormStatusActive: reachable at its URL and shown to the organizer
//   - FormStatusArchived: effectively deleted, but restorable from the
//     forms dashboard
type FormStatus string

// FormStatus values.
const (
	FormStatusDraft    FormStatus = "draft"
	FormStatusActive   FormStatus = "active"
	FormStatusArchived FormStatus = "archived"
)

// SubmitButtonWidth controls how the submit button sizes itself: shrink to
// fit its label ("auto") or fill its container ("fill").
type SubmitButtonWidth string

// SubmitButtonWidth values.
const (
	SubmitButtonWidthAuto SubmitButtonWidth = "auto"
	SubmitButtonWidthFill SubmitButtonWidth = "fill"
)

// SubmitButtonAlign controls the horizontal alignment of an "auto"-width
// submit button.
type SubmitButtonAlign string

// SubmitButtonAlign values.
const (
	SubmitButtonAlignLeft   SubmitButtonAlign = "left"
	SubmitButtonAlignCenter SubmitButtonAlign = "center"
)

// ElementType is the kind of a form element (a question or a static block).
type ElementType string

// ElementType values. Unknown values from a future API version will surface
// as the raw string.
const (
	ElementTypeText                    ElementType = "Text"
	ElementTypeMultiLineText           ElementType = "Multi-line Text"
	ElementTypeSelectOne               ElementType = "Select One"
	ElementTypeSelectOneOrOther        ElementType = `Select One or "Other"`
	ElementTypeSelectMultiple          ElementType = "Select Multiple"
	ElementTypeSelectMultipleWithOther ElementType = `Select Multiple with "Other"`
	ElementTypeYesNo                   ElementType = "Yes/No"
	ElementTypeDate                    ElementType = "Date"
	ElementTypeFile                    ElementType = "File"
	ElementTypeEmail                   ElementType = "Email"
	ElementTypePhone                   ElementType = "Phone"
	ElementTypeStaticText              ElementType = "Static Text"
	ElementTypeStaticImage             ElementType = "Static Image"
)

// LogicOperator is how option indices are combined in a selection-element
// logic group.
type LogicOperator string

// LogicOperator values.
const (
	LogicOperatorAllOf  LogicOperator = "all_of"
	LogicOperatorAnyOf  LogicOperator = "any_of"
	LogicOperatorNoneOf LogicOperator = "none_of"
)

// LogicSelectionType is the comparison used in a Yes/No-element logic group.
type LogicSelectionType string

// LogicSelectionType values.
const (
	LogicSelectionTypeIs    LogicSelectionType = "is"
	LogicSelectionTypeIsNot LogicSelectionType = "is_not"
)

// ObjectType is the object a custom property is stored against.
type ObjectType string

// ObjectType values.
const (
	ObjectTypeContact ObjectType = "contact"
	ObjectTypeOrder   ObjectType = "order"
	ObjectTypeTicket  ObjectType = "ticket"
)

// MoveSection is the section an order-form element is moved into when
// repositioned.
type MoveSection string

// MoveSection values.
const (
	MoveSectionBuyer        MoveSection = "buyer"
	MoveSectionTicketHolder MoveSection = "ticket-holder"
)

// FileAcceptType is a category of file a File element accepts.
type FileAcceptType string

// FileAcceptType values.
const (
	FileAcceptImages    FileAcceptType = "images"
	FileAcceptDocuments FileAcceptType = "documents"
	FileAcceptData      FileAcceptType = "data"
	FileAcceptAudio     FileAcceptType = "audio"
	FileAcceptVideo     FileAcceptType = "video"
)

// ForEventItemType is the kind of event item an order-form element is scoped
// to.
type ForEventItemType string

// ForEventItemType values.
const (
	ForEventItemTypeEventItem       ForEventItemType = "eventItem"
	ForEventItemTypeEventProduct    ForEventItemType = "eventProduct"
	ForEventItemTypeCheckoutProduct ForEventItemType = "checkoutProduct"
)

// FormSummary is the compact projection returned by [Client.List] and
// [Client.ListAutoPaginate]. The full element tree is omitted; use
// [Client.Retrieve] for the complete [Form].
type FormSummary struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	NumElements int        `json:"numElements"`
	Type        FormType   `json:"type"`
	Status      FormStatus `json:"status"`
}

// Form is the full form shape returned by the detail and mutation endpoints.
// It is the union of the standalone and order-form projections; order-only
// fields (such as AttendeeRowsStart, and the per-element ForEventItems /
// AskAllAttendees) are populated only for [FormTypeOrder] forms.
type Form struct {
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	NameHidden        bool              `json:"nameHidden,omitempty"`
	OwnerID           string            `json:"ownerId"`
	Status            FormStatus        `json:"status"`
	Rows              []Row             `json:"rows"`
	SubmitButtonText  string            `json:"submitButtonText"`
	SubmitButtonWidth SubmitButtonWidth `json:"submitButtonWidth"`
	SubmitButtonAlign SubmitButtonAlign `json:"submitButtonAlign,omitempty"`
	Type              FormType          `json:"type"`
	Elements          []Element         `json:"elements"`

	// AttendeeRowsStart marks the row index at which per-attendee questions
	// begin on an order form. Nil on standalone forms.
	AttendeeRowsStart *int `json:"attendeeRowsStart,omitempty"`
}

// Row is one row of the form layout, holding one or more columns.
type Row struct {
	Columns []Column `json:"columns"`
}

// Column places a single element (by its index in [Form.Elements]) into a
// row at a given fractional width.
type Column struct {
	ElementIndex  int     `json:"elementIndex"`
	WidthFraction float64 `json:"widthFraction"`
}

// Element is a single form element. It is the union of every element type;
// which fields are populated depends on [Element.Type]. For example, Options
// is set for the selection types, OtherPrompt for the "...or Other" types,
// Content for Static Text, and ImageURL for Static Image.
type Element struct {
	ID           string           `json:"id,omitempty"`
	Prompt       string           `json:"prompt,omitempty"`
	PromptHidden bool             `json:"promptHidden,omitempty"`
	HelperText   string           `json:"helperText,omitempty"`
	Type         ElementType      `json:"type"`
	Required     bool             `json:"required,omitempty"`
	PropertyID   string           `json:"propertyId,omitempty"`
	PropertyData *PropertyData    `json:"propertyData,omitempty"`
	ContactField string           `json:"contactField,omitempty"`
	Placeholder  string           `json:"placeholder,omitempty"`
	Options      []string         `json:"options,omitempty"`
	Dropdown     *bool            `json:"dropdown,omitempty"`
	OtherPrompt  string           `json:"otherPrompt,omitempty"`
	MinChoices   *int             `json:"minChoices,omitempty"`
	MaxChoices   *int             `json:"maxChoices,omitempty"`
	Min          string           `json:"min,omitempty"`
	Max          string           `json:"max,omitempty"`
	Accept       []FileAcceptType `json:"accept,omitempty"`
	Content      string           `json:"content,omitempty"`
	ImageURL     string           `json:"imageUrl,omitempty"`
	ImageWidth   *float64         `json:"imageWidth,omitempty"`
	LogicGroups  []LogicGroup     `json:"logicGroups,omitempty"`

	// ForEventItems and AskAllAttendees apply only to order-form elements.
	ForEventItems   []ForEventItem `json:"forEventItems,omitempty"`
	AskAllAttendees *bool          `json:"askAllAttendees,omitempty"`
}

// PropertyData describes the custom property an element is backed by.
type PropertyData struct {
	Type       ElementType      `json:"type"`
	ObjectType ObjectType       `json:"objectType"`
	Status     string           `json:"status"`
	Options    []PropertyOption `json:"options,omitempty"`
}

// PropertyOption is one value/label pair for a selection-backed property.
type PropertyOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// LogicGroup is a single conditional-visibility rule on an element. It is the
// union of the selection-element shape (OptionIndices + Operator) and the
// Yes/No-element shape (SelectionType + Value); RevealedElementIndex is set on
// both.
type LogicGroup struct {
	RevealedElementIndex int                `json:"revealedElementIndex"`
	OptionIndices        []int              `json:"optionIndices,omitempty"`
	Operator             LogicOperator      `json:"operator,omitempty"`
	SelectionType        LogicSelectionType `json:"selectionType,omitempty"`
	Value                *bool              `json:"value,omitempty"`
}

// ForEventItem scopes an order-form element to a specific event item,
// event product, or checkout product. Which id fields are populated depends
// on Type.
type ForEventItem struct {
	Type       ForEventItemType `json:"type"`
	EventID    string           `json:"eventId,omitempty"`
	ItemID     string           `json:"itemId,omitempty"`
	ProductID  string           `json:"productId,omitempty"`
	CheckoutID string           `json:"checkoutId,omitempty"`
}

// ListParams are the query parameters accepted by [Client.List] and
// [Client.ListAutoPaginate].
type ListParams struct {
	// Page is the 0-indexed page number. Nil uses the server default (0).
	Page *int

	// PageSize is the items-per-page cap. Nil uses the server default (50);
	// the server caps at 250.
	PageSize *int

	// Type restricts the results to a single form type. Empty fetches all
	// types.
	Type FormType
}

// CreateParams is the body accepted by [Client.Create]. Name and Type are
// required; the rest default server-side.
type CreateParams struct {
	Name              string            `json:"name"`
	NameHidden        *bool             `json:"nameHidden,omitempty"`
	Status            FormStatus        `json:"status,omitempty"`
	SubmitButtonText  string            `json:"submitButtonText,omitempty"`
	SubmitButtonWidth SubmitButtonWidth `json:"submitButtonWidth,omitempty"`
	SubmitButtonAlign SubmitButtonAlign `json:"submitButtonAlign,omitempty"`
	Type              FormType          `json:"type"`
}

// UpdateParams is the body accepted by [Client.Update]. Every field is
// optional; only the fields you set are changed (a partial update).
type UpdateParams struct {
	Name              *string           `json:"name,omitempty"`
	NameHidden        *bool             `json:"nameHidden,omitempty"`
	Status            FormStatus        `json:"status,omitempty"`
	SubmitButtonText  *string           `json:"submitButtonText,omitempty"`
	SubmitButtonWidth SubmitButtonWidth `json:"submitButtonWidth,omitempty"`
	SubmitButtonAlign SubmitButtonAlign `json:"submitButtonAlign,omitempty"`
}

// DuplicateParams is the body accepted by [Client.Duplicate]. Both fields are
// optional; the copy inherits the source's values when they are omitted.
type DuplicateParams struct {
	Name   string     `json:"name,omitempty"`
	Status FormStatus `json:"status,omitempty"`
}

// AddElementParams is the body accepted by [Client.AddElement]. Prompt and
// Type are required; the remaining fields apply to specific element types
// (see [Element] for which fields belong to which type).
type AddElementParams struct {
	Prompt       string           `json:"prompt"`
	Type         ElementType      `json:"type"`
	PromptHidden *bool            `json:"promptHidden,omitempty"`
	HelperText   string           `json:"helperText,omitempty"`
	Required     *bool            `json:"required,omitempty"`
	PropertyID   string           `json:"propertyId,omitempty"`
	ContactField string           `json:"contactField,omitempty"`
	Placeholder  string           `json:"placeholder,omitempty"`
	Options      []string         `json:"options,omitempty"`
	Dropdown     *bool            `json:"dropdown,omitempty"`
	OtherPrompt  string           `json:"otherPrompt,omitempty"`
	MinChoices   *int             `json:"minChoices,omitempty"`
	MaxChoices   *int             `json:"maxChoices,omitempty"`
	Min          string           `json:"min,omitempty"`
	Max          string           `json:"max,omitempty"`
	Accept       []FileAcceptType `json:"accept,omitempty"`
	Content      string           `json:"content,omitempty"`
	ImageURL     string           `json:"imageUrl,omitempty"`
	ImageWidth   *float64         `json:"imageWidth,omitempty"`

	// ForEventItems and AskAllAttendees apply only to order-form elements.
	ForEventItems   []ForEventItem `json:"forEventItems,omitempty"`
	AskAllAttendees *bool          `json:"askAllAttendees,omitempty"`
}

// UpdateElementParams is the body accepted by [Client.UpdateElement]. Every
// field is optional; only the fields you set are changed.
type UpdateElementParams struct {
	Prompt        *string          `json:"prompt,omitempty"`
	PromptHidden  *bool            `json:"promptHidden,omitempty"`
	HelperText    *string          `json:"helperText,omitempty"`
	Placeholder   *string          `json:"placeholder,omitempty"`
	Required      *bool            `json:"required,omitempty"`
	PropertyID    *string          `json:"propertyId,omitempty"`
	ContactField  *string          `json:"contactField,omitempty"`
	Options       []string         `json:"options,omitempty"`
	Dropdown      *bool            `json:"dropdown,omitempty"`
	OtherPrompt   *string          `json:"otherPrompt,omitempty"`
	MinChoices    *int             `json:"minChoices,omitempty"`
	MaxChoices    *int             `json:"maxChoices,omitempty"`
	Min           *string          `json:"min,omitempty"`
	Max           *string          `json:"max,omitempty"`
	Accept        []FileAcceptType `json:"accept,omitempty"`
	LogicGroups   []LogicGroup     `json:"logicGroups,omitempty"`
	Content       *string          `json:"content,omitempty"`
	ImageURL      *string          `json:"imageUrl,omitempty"`
	ImageWidth    *float64         `json:"imageWidth,omitempty"`
	ForEventItems []ForEventItem   `json:"forEventItems,omitempty"`
}

// MoveElementParams is the body accepted by [Client.MoveElement]. Position is
// the target index; Section is only meaningful on order forms.
type MoveElementParams struct {
	Position int         `json:"position"`
	Section  MoveSection `json:"section,omitempty"`
}

// EnableOtherOptionParams is the body accepted by [Client.EnableOtherOption].
type EnableOtherOptionParams struct {
	OtherPrompt string `json:"otherPrompt"`
}

// LogicCondition is the condition half of an [AddLogicRuleParams]. It is the
// union of the selection-element shape (OptionIndices + Operator) and the
// Yes/No-element shape (SelectionType + Value).
type LogicCondition struct {
	OptionIndices []int              `json:"optionIndices,omitempty"`
	Operator      LogicOperator      `json:"operator,omitempty"`
	SelectionType LogicSelectionType `json:"selectionType,omitempty"`
	Value         *bool              `json:"value,omitempty"`
}

// AddLogicRuleParams is the body accepted by [Client.AddLogicRule].
// RevealedElementID is the element shown when Condition is satisfied.
type AddLogicRuleParams struct {
	RevealedElementID string         `json:"revealedElementId"`
	Condition         LogicCondition `json:"condition"`
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

// formEnvelope is the {"data": Form} shape used by detail-returning endpoints.
type formEnvelope struct {
	Data Form `json:"data"`
}

// elementEnvelope is the {"data": Element} shape used by element-returning
// endpoints.
type elementEnvelope struct {
	Data Element `json:"data"`
}

// deleteElementEnvelope is the {"data": {deletedElementId}} shape used by
// [Client.DeleteElement].
type deleteElementEnvelope struct {
	Data DeleteElementResult `json:"data"`
}
