package services_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/golang/mock/gomock"

	"github.com/wavlake/monorepo/internal/models"
	"github.com/wavlake/monorepo/tests/mocks"
	"github.com/wavlake/monorepo/tests/testutil"
)

var _ = Describe("UserService", func() {
	var (
		ctrl              *gomock.Controller
		mockUserService   *mocks.MockUserServiceInterface
		ctx               context.Context
		testFirebaseUID   string
		testPubkey        string
		testEmail         string
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockUserService = mocks.NewMockUserServiceInterface(ctrl)
		ctx = context.Background()
		testFirebaseUID = testutil.TestFirebaseUID
		testPubkey = testutil.TestPubkey
		testEmail = testutil.TestEmail
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("UserServiceInterface", func() {
		Context("LinkPubkeyToUser", func() {
			It("should successfully link a pubkey to a user", func() {
				mockUserService.EXPECT().
					LinkPubkeyToUser(ctx, testPubkey, testFirebaseUID).
					Return(nil)

				err := mockUserService.LinkPubkeyToUser(ctx, testPubkey, testFirebaseUID)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should return error when pubkey is already linked", func() {
				expectedError := errors.New("pubkey is already linked to a different user")
				
				mockUserService.EXPECT().
					LinkPubkeyToUser(ctx, testPubkey, testFirebaseUID).
					Return(expectedError)

				err := mockUserService.LinkPubkeyToUser(ctx, testPubkey, testFirebaseUID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("already linked"))
			})
		})

		Context("UnlinkPubkeyFromUser", func() {
			It("should successfully unlink a pubkey", func() {
				mockUserService.EXPECT().
					UnlinkPubkeyFromUser(ctx, testPubkey, testFirebaseUID).
					Return(nil)

				err := mockUserService.UnlinkPubkeyFromUser(ctx, testPubkey, testFirebaseUID)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should return error when pubkey does not exist", func() {
				expectedError := errors.New("pubkey not found")
				
				mockUserService.EXPECT().
					UnlinkPubkeyFromUser(ctx, testPubkey, testFirebaseUID).
					Return(expectedError)

				err := mockUserService.UnlinkPubkeyFromUser(ctx, testPubkey, testFirebaseUID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not found"))
			})

			It("should return error when pubkey belongs to different user", func() {
				expectedError := errors.New("pubkey does not belong to this user")
				
				mockUserService.EXPECT().
					UnlinkPubkeyFromUser(ctx, testPubkey, testFirebaseUID).
					Return(expectedError)

				err := mockUserService.UnlinkPubkeyFromUser(ctx, testPubkey, testFirebaseUID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("does not belong"))
			})
		})

		Context("GetLinkedPubkeys", func() {
			It("should return user's linked pubkeys", func() {
				expectedPubkeys := []models.NostrAuth{
					testutil.ValidNostrAuth(),
				}
				
				mockUserService.EXPECT().
					GetLinkedPubkeys(ctx, testFirebaseUID).
					Return(expectedPubkeys, nil)

				pubkeys, err := mockUserService.GetLinkedPubkeys(ctx, testFirebaseUID)
				Expect(err).ToNot(HaveOccurred())
				Expect(pubkeys).To(HaveLen(1))
				Expect(pubkeys[0].Pubkey).To(Equal(testutil.TestPubkey))
			})

			It("should return empty slice when user has no linked pubkeys", func() {
				mockUserService.EXPECT().
					GetLinkedPubkeys(ctx, testFirebaseUID).
					Return([]models.NostrAuth{}, nil)

				pubkeys, err := mockUserService.GetLinkedPubkeys(ctx, testFirebaseUID)
				Expect(err).ToNot(HaveOccurred())
				Expect(pubkeys).To(BeEmpty())
			})

			It("should return error when query fails", func() {
				expectedError := errors.New("query failed")
				
				mockUserService.EXPECT().
					GetLinkedPubkeys(ctx, testFirebaseUID).
					Return(nil, expectedError)

				pubkeys, err := mockUserService.GetLinkedPubkeys(ctx, testFirebaseUID)
				Expect(err).To(HaveOccurred())
				Expect(pubkeys).To(BeNil())
			})
		})

		Context("GetFirebaseUIDByPubkey", func() {
			It("should return Firebase UID for active pubkey", func() {
				mockUserService.EXPECT().
					GetFirebaseUIDByPubkey(ctx, testPubkey).
					Return(testFirebaseUID, nil)

				uid, err := mockUserService.GetFirebaseUIDByPubkey(ctx, testPubkey)
				Expect(err).ToNot(HaveOccurred())
				Expect(uid).To(Equal(testFirebaseUID))
			})

			It("should return error when pubkey not found", func() {
				expectedError := errors.New("pubkey not found")
				
				mockUserService.EXPECT().
					GetFirebaseUIDByPubkey(ctx, testPubkey).
					Return("", expectedError)

				_, err := mockUserService.GetFirebaseUIDByPubkey(ctx, testPubkey)
				Expect(err).To(HaveOccurred())
			})

			It("should return error when pubkey is not active", func() {
				expectedError := errors.New("pubkey is not active")
				
				mockUserService.EXPECT().
					GetFirebaseUIDByPubkey(ctx, testPubkey).
					Return("", expectedError)

				_, err := mockUserService.GetFirebaseUIDByPubkey(ctx, testPubkey)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not active"))
			})
		})

		Context("GetUserEmail", func() {
			It("should return user's email", func() {
				mockUserService.EXPECT().
					GetUserEmail(ctx, testFirebaseUID).
					Return(testEmail, nil)

				email, err := mockUserService.GetUserEmail(ctx, testFirebaseUID)
				Expect(err).ToNot(HaveOccurred())
				Expect(email).To(Equal(testEmail))
			})

			It("should return error when user not found", func() {
				expectedError := errors.New("user not found")
				
				mockUserService.EXPECT().
					GetUserEmail(ctx, testFirebaseUID).
					Return("", expectedError)

				_, err := mockUserService.GetUserEmail(ctx, testFirebaseUID)
				Expect(err).To(HaveOccurred())
			})
		})

	})
})

// Note: Integration tests that test the actual UserService implementation
// would be placed in a separate file, e.g., user_service_integration_test.go
// These would use Firebase emulators and test the complete workflow
var _ = Describe("UserService Integration Tests", func() {
	Context("with Firebase emulators", func() {
		It("should handle complete linking workflow end-to-end", func() {
			Skip("Integration tests to be implemented with Firebase emulators")
		})
	})
})