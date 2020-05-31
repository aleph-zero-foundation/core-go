package tcp_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"sync"
	"time"

	"github.com/rs/zerolog"

	. "gitlab.com/alephledger/core-go/pkg/network/tcp"
)

var _ = Describe("tcp", func() {
	Context("interrupted connection", func() {
		It("should be interrupted and return an error", func() {
			serv1 := NewServer("localhost:6666", []string{"localhost:6667"}, zerolog.Nop())
			err1 := serv1.Start()
			Expect(err1).NotTo(HaveOccurred())
			defer serv1.Stop()

			serv2 := NewServer("localhost:6667", []string{"localhost:6666"}, zerolog.Nop())
			err2 := serv2.Start()
			Expect(err2).NotTo(HaveOccurred())
			defer serv2.Stop()

			con, err3 := serv1.Dial(0)
			Expect(err3).NotTo(HaveOccurred())

			buffer := make([]byte, 1)
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err4 := con.Read(buffer)
				Expect(err4).To(HaveOccurred())
			}()
			wg.Add(1)
			go func() {
				defer wg.Done()
				// should be enough
				time.Sleep(time.Second)
				err := con.Interrupt()
				Expect(err).NotTo(HaveOccurred())
			}()
			wg.Wait()
		})

		It("should be still usable", func() {
			serv1 := NewServer("localhost:6666", []string{"localhost:6667"}, zerolog.Nop())
			err1 := serv1.Start()
			Expect(err1).NotTo(HaveOccurred())
			defer serv1.Stop()

			serv2 := NewServer("localhost:6667", []string{"localhost:6666"}, zerolog.Nop())
			err2 := serv2.Start()
			Expect(err2).NotTo(HaveOccurred())
			defer serv2.Stop()

			data := []byte{1, 2, 3, 4}
			received := make([]byte, 4)
			var wg sync.WaitGroup
			interrupted := make(chan struct{})
			wg.Add(1)
			// writer
			go func() {
				defer wg.Done()
				con, err := serv1.Listen()
				Expect(err).NotTo(HaveOccurred())

				<-interrupted

				n, err := con.Write(append([]byte{}, data...))
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(len(data)))
				err = con.Close()
				Expect(err).NotTo(HaveOccurred())
			}()

			wg.Add(1)
			// reader
			go func() {
				defer wg.Done()
				con, err := serv2.Dial(0)
				Expect(err).NotTo(HaveOccurred())

				var lwg sync.WaitGroup
				lwg.Add(1)
				go func() {
					defer close(interrupted)
					defer lwg.Done()
					_, err := con.Read(received)
					Expect(err).To(HaveOccurred())
				}()
				// interrupt
				lwg.Add(1)
				go func() {
					defer lwg.Done()
					// it should be enough
					time.Sleep(time.Second)
					con.Interrupt()
				}()

				lwg.Wait()

				n, err := con.Read(received)
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(len(data)))
				err = con.Close()
				Expect(err).NotTo(HaveOccurred())
			}()
			wg.Wait()
			Expect(received).To(Equal(data))
		})
	})

	Context("not interrupted connection", func() {
		Context("listening process is writing", func() {
			It(" should finish normally", func() {
				serv1 := NewServer("localhost:6666", []string{"localhost:6667"}, zerolog.Nop())
				err1 := serv1.Start()
				Expect(err1).NotTo(HaveOccurred())
				defer serv1.Stop()

				serv2 := NewServer("localhost:6667", []string{"localhost:6666"}, zerolog.Nop())
				err2 := serv2.Start()
				Expect(err2).NotTo(HaveOccurred())
				defer serv2.Stop()

				data := []byte{1, 2, 3, 4}
				received := make([]byte, 4)
				var wg sync.WaitGroup
				wg.Add(1)
				go func() {
					defer wg.Done()
					con, err := serv1.Listen()
					Expect(err).NotTo(HaveOccurred())
					n, err := con.Write(append([]byte{}, data...))
					Expect(err).NotTo(HaveOccurred())
					Expect(n).To(Equal(len(data)))
					err = con.Close()
					Expect(err).NotTo(HaveOccurred())
				}()
				wg.Add(1)
				go func() {
					defer wg.Done()
					con, err := serv2.Dial(0)
					Expect(err).NotTo(HaveOccurred())
					n, err := con.Read(received)
					Expect(err).NotTo(HaveOccurred())
					Expect(n).To(Equal(len(data)))
					err = con.Close()
					Expect(err).NotTo(HaveOccurred())
				}()
				wg.Wait()
				Expect(received).To(Equal(data))
			})
		})

		Context("dialing process is writing", func() {
			It("should finish normally", func() {
				serv1 := NewServer("localhost:6666", []string{"localhost:6667"}, zerolog.Nop())
				err1 := serv1.Start()
				Expect(err1).NotTo(HaveOccurred())
				defer serv1.Stop()

				serv2 := NewServer("localhost:6667", []string{"localhost:6666"}, zerolog.Nop())
				err2 := serv2.Start()
				Expect(err2).NotTo(HaveOccurred())
				defer serv2.Stop()

				data := []byte{1, 2, 3, 4}
				received := make([]byte, 4)
				var wg sync.WaitGroup
				wg.Add(1)
				go func() {
					defer wg.Done()
					con, err := serv1.Listen()
					Expect(err).NotTo(HaveOccurred())
					n, err := con.Read(received)
					Expect(err).NotTo(HaveOccurred())
					Expect(n).To(Equal(len(data)))
					err = con.Close()
					Expect(err).NotTo(HaveOccurred())
				}()
				wg.Add(1)
				go func() {
					defer wg.Done()
					con, err := serv2.Dial(0)
					Expect(err).NotTo(HaveOccurred())
					n, err := con.Write(append([]byte{}, data...))
					Expect(err).NotTo(HaveOccurred())
					Expect(n).To(Equal(len(data)))
					err = con.Close()
					Expect(err).NotTo(HaveOccurred())
				}()
				wg.Wait()
				Expect(received).To(Equal(data))
			})
		})
	})

})
