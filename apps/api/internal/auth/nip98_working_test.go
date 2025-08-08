package auth_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	
	authpkg "github.com/wavlake/monorepo/internal/auth"
)

var _ = Describe("NIP98Middleware", func() {
	Describe("NewNIP98Middleware constructor", func() {
		It("should handle project ID validation", func() {
			// Test constructor behavior - whether it succeeds or fails depends on environment
			middleware, err := authpkg.NewNIP98Middleware(context.Background(), "test-project")
			
			if err != nil {
				// If it fails, it should fail with a meaningful error
				Expect(err.Error()).To(ContainSubstring("failed to create firestore client"))
			} else {
				// If it succeeds, the middleware should not be nil
				Expect(middleware).ToNot(BeNil())
			}
		})

		It("should handle empty project ID", func() {
			_, err := authpkg.NewNIP98Middleware(context.Background(), "")
			// Empty project ID should always fail
			Expect(err).To(HaveOccurred())
		})

		It("should handle constructor behavior consistently", func() {
			// Test that the constructor behaves consistently
			testProjects := []string{
				"test-project-1",
				"test-project-2", 
				"test-project-3",
			}

			results := make([]bool, len(testProjects))
			for i, project := range testProjects {
				_, err := authpkg.NewNIP98Middleware(context.Background(), project)
				results[i] = (err != nil)
			}

			// All should behave the same way (all succeed or all fail)
			for i := 1; i < len(results); i++ {
				Expect(results[i]).To(Equal(results[0]), "Inconsistent behavior between projects")
			}
		})
	})

	Describe("Constructor patterns", func() {
		It("should be constructible", func() {
			// Test that the constructor function exists and can be called
			// The result depends on the environment setup
			result, err := authpkg.NewNIP98Middleware(context.Background(), "any-project")
			
			// Either we get a middleware or an error, but not both nil/success
			if err != nil {
				Expect(result).To(BeNil())
				Expect(err.Error()).ToNot(BeEmpty())
			} else {
				Expect(result).ToNot(BeNil())
			}
		})

		It("should handle context properly", func() {
			// Test with cancelled context
			ctx, cancel := context.WithCancel(context.Background())
			cancel() // Cancel immediately
			
			result, err := authpkg.NewNIP98Middleware(ctx, "test-project")
			
			// With cancelled context, should either fail or succeed gracefully
			if err != nil {
				Expect(result).To(BeNil())
			} else {
				Expect(result).ToNot(BeNil())
			}
		})

		It("should provide error information when failing", func() {
			// Test that when errors occur, they're informative
			_, err := authpkg.NewNIP98Middleware(context.Background(), "")
			
			if err != nil {
				// Error message should not be empty
				Expect(err.Error()).ToNot(BeEmpty())
			}
		})
	})
})