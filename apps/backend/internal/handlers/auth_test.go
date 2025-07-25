package handlers_test

import (
	"errors"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/golang/mock/gomock"

	"github.com/wavlake/monorepo/internal/handlers"
	"github.com/wavlake/monorepo/internal/models"
	"github.com/wavlake/monorepo/tests/mocks"
	"github.com/wavlake/monorepo/tests/testutil"
)

var _ = Describe("AuthHandlers", func() {
	var (
		ctrl            *gomock.Controller
		mockUserService *mocks.MockUserServiceInterface
		authHandlers    *handlers.AuthHandlers
		testFirebaseUID string
		testPubkey      string
		testEmail       string
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockUserService = mocks.NewMockUserServiceInterface(ctrl)
		authHandlers = handlers.NewAuthHandlers(mockUserService)
		testFirebaseUID = testutil.TestFirebaseUID
		testPubkey = testutil.TestPubkey
		testEmail = testutil.TestEmail
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("LinkPubkey", func() {
		Context("when dual authentication is present", func() {
			It("should successfully link pubkey", func() {
				c, w := testutil.SetupGinTestContext("POST", "/v1/auth/link-pubkey", 
					testutil.ValidLinkPubkeyRequest())
				testutil.SetAuthContext(c, testFirebaseUID, testPubkey)

				mockUserService.EXPECT().
					LinkPubkeyToUser(c.Request.Context(), testPubkey, testFirebaseUID).
					Return(nil)

				authHandlers.LinkPubkey(c)

				Expect(w.Code).To(Equal(http.StatusOK))
				response := testutil.AssertJSONResponse(w, http.StatusOK)
				Expect(response["success"]).To(BeTrue())
				Expect(response["message"]).To(Equal("Pubkey linked successfully to Firebase account"))
				Expect(response["firebase_uid"]).To(Equal(testFirebaseUID))
				Expect(response["pubkey"]).To(Equal(testPubkey))
				Expect(response["linked_at"]).ToNot(BeEmpty())
			})

			It("should validate request pubkey matches authenticated pubkey", func() {
				c, w := testutil.SetupGinTestContext("POST", "/v1/auth/link-pubkey", map[string]interface{}{
					"pubkey": "different-pubkey",
				})
				testutil.SetAuthContext(c, testFirebaseUID, testPubkey)

				authHandlers.LinkPubkey(c)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
				response := testutil.AssertJSONResponse(w, http.StatusBadRequest)
				Expect(response["error"]).To(Equal("Request pubkey does not match authenticated pubkey"))
			})
		})

		Context("when Firebase authentication is missing", func() {
			It("should return unauthorized error", func() {
				c, w := testutil.SetupGinTestContext("POST", "/v1/auth/link-pubkey", 
					testutil.ValidLinkPubkeyRequest())
				testutil.SetAuthContext(c, "", testPubkey) // Missing Firebase UID

				authHandlers.LinkPubkey(c)

				Expect(w.Code).To(Equal(http.StatusUnauthorized))
				response := testutil.AssertJSONResponse(w, http.StatusUnauthorized)
				Expect(response["error"]).To(Equal("Missing Firebase authentication"))
			})
		})

		Context("when Nostr authentication is missing", func() {
			It("should return unauthorized error", func() {
				c, w := testutil.SetupGinTestContext("POST", "/v1/auth/link-pubkey", 
					testutil.ValidLinkPubkeyRequest())
				testutil.SetAuthContext(c, testFirebaseUID, "") // Missing Nostr pubkey

				authHandlers.LinkPubkey(c)

				Expect(w.Code).To(Equal(http.StatusUnauthorized))
				response := testutil.AssertJSONResponse(w, http.StatusUnauthorized)
				Expect(response["error"]).To(Equal("Missing Nostr authentication"))
			})
		})

		Context("when service returns error", func() {
			It("should return bad request with error message", func() {
				c, w := testutil.SetupGinTestContext("POST", "/v1/auth/link-pubkey", 
					testutil.ValidLinkPubkeyRequest())
				testutil.SetAuthContext(c, testFirebaseUID, testPubkey)

				expectedError := errors.New("pubkey is already linked to a different user")
				mockUserService.EXPECT().
					LinkPubkeyToUser(c.Request.Context(), testPubkey, testFirebaseUID).
					Return(expectedError)

				authHandlers.LinkPubkey(c)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
				response := testutil.AssertJSONResponse(w, http.StatusBadRequest)
				Expect(response["error"]).To(Equal(expectedError.Error()))
			})
		})

		Context("when request body contains malformed JSON", func() {
			It("should handle malformed JSON gracefully and continue with authentication", func() {
				c, w := testutil.SetupGinTestContext("POST", "/v1/auth/link-pubkey", "{malformed json")
				testutil.SetAuthContext(c, testFirebaseUID, testPubkey)

				mockUserService.EXPECT().
					LinkPubkeyToUser(c.Request.Context(), testPubkey, testFirebaseUID).
					Return(nil)

				authHandlers.LinkPubkey(c)

				Expect(w.Code).To(Equal(http.StatusOK))
				response := testutil.AssertJSONResponse(w, http.StatusOK)
				Expect(response["success"]).To(BeTrue())
			})
		})

		Context("when auth context has wrong types", func() {
			It("should handle firebase_uid type assertion panic", func() {
				c, _ := testutil.SetupGinTestContext("POST", "/v1/auth/link-pubkey", 
					testutil.ValidLinkPubkeyRequest())
				c.Set("firebase_uid", 12345) // Wrong type - should be string
				c.Set("nostr_pubkey", testPubkey)

				Expect(func() {
					authHandlers.LinkPubkey(c)
				}).To(Panic())
			})

			It("should handle nostr_pubkey type assertion panic", func() {
				c, _ := testutil.SetupGinTestContext("POST", "/v1/auth/link-pubkey", 
					testutil.ValidLinkPubkeyRequest())
				c.Set("firebase_uid", testFirebaseUID)
				c.Set("nostr_pubkey", 12345) // Wrong type - should be string

				Expect(func() {
					authHandlers.LinkPubkey(c)
				}).To(Panic())
			})
		})

		Context("when request body has empty pubkey", func() {
			It("should successfully link without validating empty request pubkey", func() {
				c, w := testutil.SetupGinTestContext("POST", "/v1/auth/link-pubkey", map[string]interface{}{
					"pubkey": "",
				})
				testutil.SetAuthContext(c, testFirebaseUID, testPubkey)

				mockUserService.EXPECT().
					LinkPubkeyToUser(c.Request.Context(), testPubkey, testFirebaseUID).
					Return(nil)

				authHandlers.LinkPubkey(c)

				Expect(w.Code).To(Equal(http.StatusOK))
				response := testutil.AssertJSONResponse(w, http.StatusOK)
				Expect(response["success"]).To(BeTrue())
			})
		})
	})

	Describe("UnlinkPubkey", func() {
		Context("when Firebase authentication is present", func() {
			It("should successfully unlink pubkey", func() {
				c, w := testutil.SetupGinTestContext("POST", "/v1/auth/unlink-pubkey", 
					testutil.ValidUnlinkPubkeyRequest())
				testutil.SetAuthContext(c, testFirebaseUID, "")

				mockUserService.EXPECT().
					UnlinkPubkeyFromUser(c.Request.Context(), testPubkey, testFirebaseUID).
					Return(nil)

				authHandlers.UnlinkPubkey(c)

				Expect(w.Code).To(Equal(http.StatusOK))
				response := testutil.AssertJSONResponse(w, http.StatusOK)
				Expect(response["success"]).To(BeTrue())
				Expect(response["message"]).To(Equal("Pubkey unlinked successfully from Firebase account"))
				Expect(response["pubkey"]).To(Equal(testPubkey))
			})
		})

		Context("when Firebase authentication is missing", func() {
			It("should return unauthorized error", func() {
				c, w := testutil.SetupGinTestContext("POST", "/v1/auth/unlink-pubkey", 
					testutil.ValidUnlinkPubkeyRequest())
				// No auth context set

				authHandlers.UnlinkPubkey(c)

				Expect(w.Code).To(Equal(http.StatusUnauthorized))
				response := testutil.AssertJSONResponse(w, http.StatusUnauthorized)
				Expect(response["error"]).To(Equal("Missing Firebase authentication"))
			})
		})

		Context("when request body is invalid", func() {
			It("should return bad request error", func() {
				c, w := testutil.SetupGinTestContext("POST", "/v1/auth/unlink-pubkey", map[string]interface{}{
					// Missing required pubkey field
				})
				testutil.SetAuthContext(c, testFirebaseUID, "")

				authHandlers.UnlinkPubkey(c)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
				response := testutil.AssertJSONResponse(w, http.StatusBadRequest)
				Expect(response["error"]).To(Equal("Invalid request body"))
			})
		})

		Context("when service returns error", func() {
			It("should return bad request with error message", func() {
				c, w := testutil.SetupGinTestContext("POST", "/v1/auth/unlink-pubkey", 
					testutil.ValidUnlinkPubkeyRequest())
				testutil.SetAuthContext(c, testFirebaseUID, "")

				expectedError := errors.New("pubkey not found")
				mockUserService.EXPECT().
					UnlinkPubkeyFromUser(c.Request.Context(), testPubkey, testFirebaseUID).
					Return(expectedError)

				authHandlers.UnlinkPubkey(c)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
				response := testutil.AssertJSONResponse(w, http.StatusBadRequest)
				Expect(response["error"]).To(Equal(expectedError.Error()))
			})
		})

		Context("when request contains malformed JSON", func() {
			It("should return bad request error for malformed JSON", func() {
				c, w := testutil.SetupGinTestContext("POST", "/v1/auth/unlink-pubkey", "{malformed json")
				testutil.SetAuthContext(c, testFirebaseUID, "")

				authHandlers.UnlinkPubkey(c)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
				response := testutil.AssertJSONResponse(w, http.StatusBadRequest)
				Expect(response["error"]).To(Equal("Invalid request body"))
			})
		})

		Context("when pubkey field is empty string", func() {
			It("should return bad request error for empty pubkey", func() {
				c, w := testutil.SetupGinTestContext("POST", "/v1/auth/unlink-pubkey", map[string]interface{}{
					"pubkey": "",
				})
				testutil.SetAuthContext(c, testFirebaseUID, "")

				authHandlers.UnlinkPubkey(c)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
				response := testutil.AssertJSONResponse(w, http.StatusBadRequest)
				Expect(response["error"]).To(Equal("Invalid request body"))
			})
		})

		Context("when Firebase UID context has wrong type", func() {
			It("should handle firebase_uid type assertion panic", func() {
				c, _ := testutil.SetupGinTestContext("POST", "/v1/auth/unlink-pubkey", 
					testutil.ValidUnlinkPubkeyRequest())
				c.Set("firebase_uid", 12345) // Wrong type - should be string

				Expect(func() {
					authHandlers.UnlinkPubkey(c)
				}).To(Panic())
			})
		})
	})

	Describe("GetLinkedPubkeys", func() {
		Context("when Firebase authentication is present", func() {
			It("should return linked pubkeys list", func() {
				c, w := testutil.SetupGinTestContext("GET", "/v1/auth/get-linked-pubkeys", nil)
				testutil.SetAuthContext(c, testFirebaseUID, "")

				expectedPubkeys := []models.NostrAuth{
					{
						Pubkey:     testPubkey,
						LinkedAt:   time.Now(),
						LastUsedAt: time.Now(),
					},
				}

				mockUserService.EXPECT().
					GetLinkedPubkeys(c.Request.Context(), testFirebaseUID).
					Return(expectedPubkeys, nil)

				authHandlers.GetLinkedPubkeys(c)

				Expect(w.Code).To(Equal(http.StatusOK))
				response := testutil.AssertJSONResponse(w, http.StatusOK)
				Expect(response["success"]).To(BeTrue())
				Expect(response["firebase_uid"]).To(Equal(testFirebaseUID))
				
				linkedPubkeys, ok := response["linked_pubkeys"].([]interface{})
				Expect(ok).To(BeTrue())
				Expect(linkedPubkeys).To(HaveLen(1))
			})

			It("should return empty array when no pubkeys linked", func() {
				c, w := testutil.SetupGinTestContext("GET", "/v1/auth/get-linked-pubkeys", nil)
				testutil.SetAuthContext(c, testFirebaseUID, "")

				mockUserService.EXPECT().
					GetLinkedPubkeys(c.Request.Context(), testFirebaseUID).
					Return([]models.NostrAuth{}, nil)

				authHandlers.GetLinkedPubkeys(c)

				Expect(w.Code).To(Equal(http.StatusOK))
				response := testutil.AssertJSONResponse(w, http.StatusOK)
				Expect(response["success"]).To(BeTrue())
				
				linkedPubkeys, ok := response["linked_pubkeys"].([]interface{})
				Expect(ok).To(BeTrue())
				Expect(linkedPubkeys).To(HaveLen(0))
			})
		})

		Context("when Firebase authentication is missing", func() {
			It("should return unauthorized error", func() {
				c, w := testutil.SetupGinTestContext("GET", "/v1/auth/get-linked-pubkeys", nil)
				// No auth context set

				authHandlers.GetLinkedPubkeys(c)

				Expect(w.Code).To(Equal(http.StatusUnauthorized))
				response := testutil.AssertJSONResponse(w, http.StatusUnauthorized)
				Expect(response["error"]).To(Equal("Missing Firebase authentication"))
			})
		})

		Context("when service returns error", func() {
			It("should return internal server error", func() {
				c, w := testutil.SetupGinTestContext("GET", "/v1/auth/get-linked-pubkeys", nil)
				testutil.SetAuthContext(c, testFirebaseUID, "")

				expectedError := errors.New("database connection failed")
				mockUserService.EXPECT().
					GetLinkedPubkeys(c.Request.Context(), testFirebaseUID).
					Return(nil, expectedError)

				authHandlers.GetLinkedPubkeys(c)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				response := testutil.AssertJSONResponse(w, http.StatusInternalServerError)
				Expect(response["error"]).To(Equal("Failed to retrieve linked pubkeys"))
				Expect(response["debug"]).To(Equal(expectedError.Error()))
			})
		})

		Context("when Firebase UID context has wrong type", func() {
			It("should handle firebase_uid type assertion panic", func() {
				c, _ := testutil.SetupGinTestContext("GET", "/v1/auth/get-linked-pubkeys", nil)
				c.Set("firebase_uid", 12345) // Wrong type - should be string

				Expect(func() {
					authHandlers.GetLinkedPubkeys(c)
				}).To(Panic())
			})
		})
	})

	Describe("CheckPubkeyLink", func() {
		Context("when Nostr authentication is present", func() {
			It("should return link status when pubkey is linked", func() {
				c, w := testutil.SetupGinTestContext("POST", "/v1/auth/check-pubkey-link", 
					testutil.ValidCheckPubkeyLinkRequest())
				testutil.SetAuthContext(c, "", testPubkey)

				mockUserService.EXPECT().
					GetFirebaseUIDByPubkey(c.Request.Context(), testPubkey).
					Return(testFirebaseUID, nil)

				mockUserService.EXPECT().
					GetUserEmail(c.Request.Context(), testFirebaseUID).
					Return(testEmail, nil)

				authHandlers.CheckPubkeyLink(c)

				Expect(w.Code).To(Equal(http.StatusOK))
				response := testutil.AssertJSONResponse(w, http.StatusOK)
				Expect(response["success"]).To(BeTrue())
				Expect(response["is_linked"]).To(BeTrue())
				Expect(response["firebase_uid"]).To(Equal(testFirebaseUID))
				Expect(response["pubkey"]).To(Equal(testPubkey))
				Expect(response["email"]).To(Equal(testEmail))
			})

			It("should return not linked when pubkey is not found", func() {
				c, w := testutil.SetupGinTestContext("POST", "/v1/auth/check-pubkey-link", 
					testutil.ValidCheckPubkeyLinkRequest())
				testutil.SetAuthContext(c, "", testPubkey)

				expectedError := errors.New("pubkey not found")
				mockUserService.EXPECT().
					GetFirebaseUIDByPubkey(c.Request.Context(), testPubkey).
					Return("", expectedError)

				authHandlers.CheckPubkeyLink(c)

				Expect(w.Code).To(Equal(http.StatusOK))
				response := testutil.AssertJSONResponse(w, http.StatusOK)
				Expect(response["success"]).To(BeTrue())
				Expect(response["is_linked"]).To(BeFalse())
				Expect(response["firebase_uid"]).To(BeNil()) // omitempty field
				Expect(response["pubkey"]).To(Equal(testPubkey))
				Expect(response["email"]).To(BeNil()) // omitempty field
			})

			It("should handle email retrieval failure gracefully", func() {
				c, w := testutil.SetupGinTestContext("POST", "/v1/auth/check-pubkey-link", 
					testutil.ValidCheckPubkeyLinkRequest())
				testutil.SetAuthContext(c, "", testPubkey)

				mockUserService.EXPECT().
					GetFirebaseUIDByPubkey(c.Request.Context(), testPubkey).
					Return(testFirebaseUID, nil)

				emailError := errors.New("email retrieval failed")
				mockUserService.EXPECT().
					GetUserEmail(c.Request.Context(), testFirebaseUID).
					Return("", emailError)

				authHandlers.CheckPubkeyLink(c)

				Expect(w.Code).To(Equal(http.StatusOK))
				response := testutil.AssertJSONResponse(w, http.StatusOK)
				Expect(response["success"]).To(BeTrue())
				Expect(response["is_linked"]).To(BeTrue())
				Expect(response["firebase_uid"]).To(Equal(testFirebaseUID))
				Expect(response["email"]).To(BeNil()) // omitempty field when empty
			})
		})

		Context("when Nostr authentication is missing", func() {
			It("should return unauthorized error", func() {
				c, w := testutil.SetupGinTestContext("POST", "/v1/auth/check-pubkey-link", 
					testutil.ValidCheckPubkeyLinkRequest())
				// No auth context set

				authHandlers.CheckPubkeyLink(c)

				Expect(w.Code).To(Equal(http.StatusUnauthorized))
				response := testutil.AssertJSONResponse(w, http.StatusUnauthorized)
				Expect(response["error"]).To(Equal("Missing Nostr authentication"))
			})
		})

		Context("when request body is invalid", func() {
			It("should return bad request error", func() {
				c, w := testutil.SetupGinTestContext("POST", "/v1/auth/check-pubkey-link", map[string]interface{}{
					// Missing required pubkey field
				})
				c.Set("pubkey", testPubkey) // Use "pubkey" not "nostr_pubkey"

				authHandlers.CheckPubkeyLink(c)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
				response := testutil.AssertJSONResponse(w, http.StatusBadRequest)
				Expect(response["error"]).To(Equal("Invalid request body - pubkey is required"))
			})
		})

		Context("when request contains malformed JSON", func() {
			It("should return bad request error for malformed JSON", func() {
				c, w := testutil.SetupGinTestContext("POST", "/v1/auth/check-pubkey-link", "{malformed json")
				c.Set("pubkey", testPubkey) // Use "pubkey" not "nostr_pubkey"

				authHandlers.CheckPubkeyLink(c)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
				response := testutil.AssertJSONResponse(w, http.StatusBadRequest)
				Expect(response["error"]).To(Equal("Invalid request body - pubkey is required"))
			})
		})

		Context("when request pubkey doesn't match authenticated pubkey", func() {
			It("should return forbidden error for pubkey mismatch", func() {
				c, w := testutil.SetupGinTestContext("POST", "/v1/auth/check-pubkey-link", map[string]interface{}{
					"pubkey": "different-pubkey-than-authenticated",
				})
				c.Set("pubkey", testPubkey) // Use "pubkey" not "nostr_pubkey"

				authHandlers.CheckPubkeyLink(c)

				Expect(w.Code).To(Equal(http.StatusForbidden))
				response := testutil.AssertJSONResponse(w, http.StatusForbidden)
				Expect(response["error"]).To(Equal("You can only check linking status for your own pubkey"))
			})
		})

		Context("when pubkey context has wrong type", func() {
			It("should handle pubkey type assertion panic", func() {
				c, _ := testutil.SetupGinTestContext("POST", "/v1/auth/check-pubkey-link", 
					testutil.ValidCheckPubkeyLinkRequest())
				c.Set("pubkey", 12345) // Wrong type - should be string

				Expect(func() {
					authHandlers.CheckPubkeyLink(c)
				}).To(Panic())
			})
		})

		Context("when authenticated pubkey does not match request pubkey", func() {
			It("should return forbidden error", func() {
				c, w := testutil.SetupGinTestContext("POST", "/v1/auth/check-pubkey-link", map[string]interface{}{
					"pubkey": "different-pubkey",
				})
				testutil.SetAuthContext(c, "", testPubkey)

				authHandlers.CheckPubkeyLink(c)

				Expect(w.Code).To(Equal(http.StatusForbidden))
				response := testutil.AssertJSONResponse(w, http.StatusForbidden)
				Expect(response["error"]).To(Equal("You can only check linking status for your own pubkey"))
			})
		})
	})
})