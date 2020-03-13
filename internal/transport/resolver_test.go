package transport_test

import (
	"github.com/hugorut/coins-oracle/pkg/transport"
	"sync"

	mock_echo "github.com/hugorut/coins-oracle/internal/handlers/mocks"

	"github.com/labstack/echo"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	. "github.com/hugorut/coins-oracle/internal/transport"
	mock_transport "github.com/hugorut/coins-oracle/internal/transport/mocks"
)

var _ = Describe("Transport", func() {
	var (
		resolver *CoinResolver
		client   transport.CoinClient
		ctrl     *gomock.Controller
		logger   echo.Logger
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		client = mock_transport.NewMockCoinClient(ctrl)
		mockLogger := mock_echo.NewMockLogger(ctrl)
		logger = mockLogger

		mockLogger.EXPECT().Print(gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Printf(gomock.Any(), gomock.Any()).AnyTimes()

		resolver = &CoinResolver{
			C:      make(map[string]transport.CoinClient),
			Mu:     &sync.Mutex{},
			Logger: logger,
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("CoinResolver", func() {
		Describe("#Register", func() {
			It("Should register a valid lambda under the namespace", func() {
				Expect(resolver.Register("test", client)).To(BeIdenticalTo(resolver))

				c, err := resolver.Get("test")

				Expect(err).ToNot(HaveOccurred())
				Expect(c).To(BeIdenticalTo(client))
			})

			It("Should use the last registered namespace", func() {
				client2 := mock_transport.NewMockCoinClient(ctrl)
				resolver.Register("test", client).Register("test", client2)

				c, err := resolver.Get("test")

				Expect(err).ToNot(HaveOccurred())
				Expect(c).To(BeIdenticalTo(client2))
			})

			It("Should error if namespace not registered", func() {
				_, err := resolver.Get("test")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not find client"))
			})
		})

		Describe("#GetNodes", func() {
			Context("With info", func() {
				It("Should return a list of registered nodes with information meta attached", func() {
					c1 := mock_transport.NewMockCoinClient(ctrl)
					c2 := mock_transport.NewMockCoinClient(ctrl)
					resolver.Register("test", c1).Register("test2", c2)

					data1 := transport.CoinData{
						Chain:        "main",
						CurrentBlock: "15",
					}
					data2 := transport.CoinData{
						Chain:        "main",
						CurrentBlock: "22",
					}

					c1.EXPECT().GetInfo().Return(&transport.CoinState{
						Data: data1,
					}, nil)
					c2.EXPECT().GetInfo().Return(&transport.CoinState{
						Data: data2,
					}, nil)
					list := resolver.GetNodes(true)

					Expect(list).To(ConsistOf(
						MatchAllFields(Fields{
							"AssetId": Equal("test"),
							"Running": Equal(true),
							"Info": PointTo(MatchFields(IgnoreExtras, Fields{
								"CurrentBlock": Equal(data1.CurrentBlock),
							})),
						}),
						MatchAllFields(Fields{
							"AssetId": Equal("test2"),
							"Running": Equal(true),
							"Info": PointTo(MatchFields(IgnoreExtras, Fields{
								"CurrentBlock": Equal(data2.CurrentBlock),
							})),
						}),
					))
				})
			})

			Context("Without info", func() {
				It("Should return a list of registered nodes", func() {
					resolver.Register("test", client).Register("test2", client).Register("test3", client)

					list := resolver.GetNodes(false)

					Expect(list).To(ConsistOf(
						MatchAllFields(Fields{
							"AssetId": Equal("test"),
							"Running": Equal(true),
							"Info":    BeNil(),
						}),
						MatchAllFields(Fields{
							"AssetId": Equal("test2"),
							"Running": Equal(true),
							"Info":    BeNil(),
						}),
						MatchAllFields(Fields{
							"AssetId": Equal("test3"),
							"Running": Equal(true),
							"Info":    BeNil(),
						}),
					))
				})
			})
		})
	})
})
