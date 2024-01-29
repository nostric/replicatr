package nip11

import (
	"sync"

	"github.com/Hubmakerlabs/replicatr/pkg/nostr/kinds"
)

type NIP struct {
	Description string
	Number      int
}

// this is the list of all nips and their titles for use in supported_nips field
var (
	BasicProtocol                  = NIP{"Basic protocol flow description", 1}
	NIP1                           = BasicProtocol
	FollowList                     = NIP{"Follow List", 2}
	NIP2                           = FollowList
	OpenTimestampsAttestations     = NIP{"OpenTimestamps Attestations for Events", 3}
	NIP3                           = OpenTimestampsAttestations
	EncryptedDirectMessage         = NIP{"Encrypted Direct Message --- unrecommended: deprecated in favor of NIP-44", 4}
	NIP4                           = EncryptedDirectMessage
	MappingNostrKeysToDNS          = NIP{"Mapping Nostr keys to DNS-based internet identifiers", 5}
	NIP5                           = MappingNostrKeysToDNS
	HandlingMentions               = NIP{"Handling Mentions --- unrecommended: deprecated in favor of NIP-27", 8}
	NIP8                           = HandlingMentions
	EventDeletion                  = NIP{"Event Deletion", 9}
	NIP9                           = EventDeletion
	RelayInformationDocument       = NIP{"Relay Information Document", 11}
	NIP11                          = RelayInformationDocument
	SubjectTag                     = NIP{"Subject tag in text events", 14}
	NIP14                          = SubjectTag
	NostrMarketplace               = NIP{"Nostr Marketplace (for resilient marketplaces)", 15}
	NIP15                          = NostrMarketplace
	Reposts                        = NIP{"Reposts", 18}
	NIP18                          = Reposts
	Bech32EncodedEntities          = NIP{"bech32-encoded entities", 19}
	NIP19                          = Bech32EncodedEntities
	NostrURIScheme                 = NIP{"nostr: URI scheme", 21}
	NIP21                          = NostrURIScheme
	LongFormContent                = NIP{"Long-form Content", 23}
	NIP23                          = LongFormContent
	ExtraMetadata                  = NIP{"Extra metadata fields and tags", 24}
	NIP24                          = ExtraMetadata
	Reactions                      = NIP{"Reactions", 25}
	NIP25                          = Reactions
	DelegatedEventSigning          = NIP{"Delegated Event Signing", 26}
	NIP26                          = DelegatedEventSigning
	TextNoteReferences             = NIP{"Text Note References", 27}
	NIP27                          = TextNoteReferences
	PublicChat                     = NIP{"Public Chat", 28}
	NIP28                          = PublicChat
	CustomEmoji                    = NIP{"Custom Emoji", 30}
	NIP30                          = CustomEmoji
	Labeling                       = NIP{"Labeling", 32}
	NIP32                          = Labeling
	SensitiveContent               = NIP{"Sensitive Content", 36}
	NIP36                          = SensitiveContent
	UserStatuses                   = NIP{"User Statuses", 38}
	NIP38                          = UserStatuses
	ExternalIdentitiesInProfiles   = NIP{"External Identities in Profiles", 39}
	NIP39                          = ExternalIdentitiesInProfiles
	ExpirationTimestamp            = NIP{"Expiration Timestamp", 40}
	NIP40                          = ExpirationTimestamp
	Authentication                 = NIP{"Authentication of clients to relays", 42}
	NIP42                          = Authentication
	VersionedEncryption            = NIP{"Versioned Encryption", 44}
	NIP44                          = VersionedEncryption
	CountingResults                = NIP{"Counting results", 45}
	NIP45                          = CountingResults
	NostrConnect                   = NIP{"Nostr Connect", 46}
	NIP46                          = NostrConnect
	WalletConnect                  = NIP{"Wallet Connect", 47}
	NIP47                          = WalletConnect
	ProxyTags                      = NIP{"Proxy Tags", 48}
	NIP48                          = ProxyTags
	SearchCapability               = NIP{"Search Capability", 50}
	NIP50                          = SearchCapability
	Lists                          = NIP{"Lists", 51}
	NIP51                          = Lists
	CalendarEvents                 = NIP{"Calendar Events", 52}
	NIP52                          = CalendarEvents
	LiveActivities                 = NIP{"Live Activities", 53}
	NIP53                          = LiveActivities
	Reporting                      = NIP{"Reporting", 56}
	NIP56                          = Reporting
	LightningZaps                  = NIP{"Lightning Zaps", 57}
	NIP57                          = LightningZaps
	Badges                         = NIP{"Badges", 58}
	NIP58                          = Badges
	RelayListMetadata              = NIP{"Relay List Metadata", 65}
	NIP65                          = RelayListMetadata
	ModeratedCommunities           = NIP{"Moderated Communities", 72}
	NIP72                          = ModeratedCommunities
	ZapGoals                       = NIP{"Zap Goals", 75}
	NIP75                          = ZapGoals
	ApplicationSpecificData        = NIP{"Application-specific data", 78}
	NIP78                          = ApplicationSpecificData
	Highlights                     = NIP{"Highlights", 84}
	NIP84                          = Highlights
	RecommendedApplicationHandlers = NIP{"Recommended Application Handlers", 89}
	NIP89                          = RecommendedApplicationHandlers
	DataVendingMachines            = NIP{"Data Vending Machines", 90}
	NIP90                          = DataVendingMachines
	FileMetadata                   = NIP{"File Metadata", 94}
	NIP94                          = FileMetadata
	HTTPFileStorageIntegration     = NIP{"HTTP File Storage Integration", 96}
	NIP96                          = HTTPFileStorageIntegration
	HTTPAuth                       = NIP{"HTTP Auth", 98}
	NIP98                          = HTTPAuth
	ClassifiedListings             = NIP{"Classified Listings", 99}
	NIP99                          = ClassifiedListings
)

type Limits struct {
	MaxMessageLength int  `json:"max_message_length,omitempty"`
	MaxSubscriptions int  `json:"max_subscriptions,omitempty"`
	MaxFilters       int  `json:"max_filters,omitempty"`
	MaxLimit         int  `json:"max_limit,omitempty"`
	MaxSubidLength   int  `json:"max_subid_length,omitempty"`
	MaxEventTags     int  `json:"max_event_tags,omitempty"`
	MaxContentLength int  `json:"max_content_length,omitempty"`
	MinPowDifficulty int  `json:"min_pow_difficulty,omitempty"`
	AuthRequired     bool `json:"auth_required"`
	PaymentRequired  bool `json:"payment_required"`
	RestrictedWrites bool `json:"restricted_writes"`
}
type Payment struct {
	Amount int    `json:"amount"`
	Unit   string `json:"unit"`
}

type Sub struct {
	Payment
	Period int `json:"period"`
}

type Pub struct {
	Kinds kinds.T `json:"kinds"`
	Payment
}

type Fees struct {
	Admission    []Payment `json:"admission,omitempty"`
	Subscription []Sub     `json:"subscription,omitempty"`
	Publication  []Pub     `json:"publication,omitempty"`
}

type NIPs []int

type Info struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	PubKey         string   `json:"pubkey"`
	Contact        string   `json:"contact"`
	Nips           NIPs     `json:"supported_nips"`
	Software       string   `json:"software"`
	Version        string   `json:"version"`
	Limitation     *Limits  `json:"limitation,omitempty"`
	RelayCountries []string `json:"relay_countries,omitempty"`
	LanguageTags   []string `json:"language_tags,omitempty"`
	Tags           []string `json:"tags,omitempty"`
	PostingPolicy  string   `json:"posting_policy,omitempty"`
	PaymentsURL    string   `json:"payments_url,omitempty"`
	Fees           *Fees    `json:"fees,omitempty"`
	Icon           string   `json:"icon"`
	sync.Mutex
}

// NewInfo populates the nips map and if an Info structure is provided it is
// used and its nips map is populated if it isn't already.
func NewInfo(inf *Info) *Info {
	if inf != nil {
		inf.Lock()
		if inf.Limitation == nil {
			inf.Limitation = &Limits{}
		}
		inf.Unlock()
		return inf
	}
	return &Info{}
}

func (inf *Info) AddNIPs(n ...int) {
	inf.Lock()
	for _, number := range n {
		inf.Nips = append(inf.Nips, number)
	}
	inf.Unlock()
}

func (inf *Info) HasNIP(n int) (ok bool) {
	inf.Lock()
	for i := range inf.Nips {
		if inf.Nips[i] == n {
			ok = true
			break
		}
	}
	inf.Unlock()
	return
}
