package checks

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/go-logr/logr"
	fastping "github.com/tatsushid/go-fastping"
)

const pingIsEnabled = false

var (
	ErrHostNotResolved    = errors.New("host_not_resolved")
	ErrHostNotAccessible  = errors.New("host_not_accessible")
	ErrHTTPRequest        = errors.New("http_request_error")
	ErrTimeoutExceed      = errors.New("timeout_exceed")
	ErrCertificateInvalid = errors.New("certificate_invalid")
)

type CheckParam struct {
	ipAddresses []net.IP
	Link        *url.URL

	Latency time.Duration
	Period  time.Duration

	EnvName string
}

func check(
	ctx context.Context,
	log logr.Logger,
	link string,
	params CheckParam,
	errCh chan<- error,
) {
	var err error

	params.Link, err = url.Parse(link)
	if err != nil {
		errCh <- ErrHostNotResolved
	}

	// check DNS
	err = resolveDNS(params.Link.Host)
	if err != nil {
		log.Error(err, "coudn't resolve domain name")
		errCh <- ErrHostNotResolved
	}

	// get ip addresses by domain name
	params.ipAddresses, err = net.LookupIP(params.Link.Host)
	if err != nil {
		log.Error(err, "coudn't look up ip")
		errCh <- ErrHostNotResolved
	}

	// ping
	if pingIsEnabled {
		for i := range params.ipAddresses {
			err = ping(params.ipAddresses[i].String())
			if err != nil {
				log.Error(err, "ping failed")
				errCh <- ErrHostNotAccessible
			}
		}
	}

	// check http client
	err = checkHTTPConnection(ctx, link)
	if err != nil {
		log.Error(err, "bad GET response")
		errCh <- ErrHTTPRequest
	}

	// check certificate's expiration date
	if params.Link.Scheme == "https" {
		err = checkCertificate(log, params.Link.Host)
		if err != nil {
			log.Error(err, "check certficate error")
			errCh <- ErrCertificateInvalid
		}
	}

	errCh <- nil
}

func resolveDNS(addr string) error {
	// nolint:ifshort // придерживаемся единого стандарта ифов
	_, err := net.LookupCNAME(addr)
	if err != nil {
		return fmt.Errorf("coudn't look up cname: %w", err)
	}

	return nil
}

func ping(ipAddr string) error {
	pinger := fastping.NewPinger()

	_, err := pinger.Network("icmp")
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	err = pinger.AddIP(ipAddr)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	err = pinger.Run()
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func checkHTTPConnection(ctx context.Context, addr string) error {
	resp, err := http.NewRequestWithContext(ctx, http.MethodGet, addr, http.NoBody)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	defer resp.Body.Close()

	return nil
}

func checkCertificate(log logr.Logger, addr string) error {
	conn, err := tls.Dial("tcp", addr+":443", nil)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	defer conn.Close()

	err = conn.Handshake()
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if len(conn.ConnectionState().PeerCertificates) == 0 {
		return errors.New("doesn't have certificate")
	}

	if conn.ConnectionState().PeerCertificates[0].NotAfter.Before(time.Now()) {
		err := errors.New("certificate has been expired")
		log.Error(err, "certificate has been expired")

		return err
	}

	return nil
}

func CheckWithTimeout(log logr.Logger, link string, params CheckParam) error {
	errCh := make(chan error)

	ctx, cancel := context.WithTimeout(context.Background(), params.Latency)
	defer cancel()

	go func() {
		check(ctx, log, link, params, errCh)
	}()

	select {
	case <-ctx.Done():
		return ErrTimeoutExceed
	case err := <-errCh:
		return err
	}
}
