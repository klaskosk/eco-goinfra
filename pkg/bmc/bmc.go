package bmc

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/golang/glog"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
	"golang.org/x/crypto/ssh"
)

const (
	defaultTimeOut = 5 * time.Second

	manufacturerDell = "Dell Inc."
	manufacturerHPE  = "HPE"
)

var (
	// DefaultTimeOuts holds the default redfish and ssh timeouts.
	DefaultTimeOuts = TimeOuts{
		Redfish: defaultTimeOut,
		SSH:     defaultTimeOut,
	}

	// CLI command to get the serial console (virtual serial port).
	cliCmdSerialConsole = map[string]string{
		manufacturerHPE:  "VSP",
		manufacturerDell: "console com2",
	}
)

// User holds the Name and Password for a user (ssh/redfish).
type User struct {
	// Name holds the user's name
	Name string
	// Password holds the user's password
	Password string
}

// TimeOuts holds the configured timeouts for Redfish and SSH acccess.
type TimeOuts struct {
	// Redfish timeout for the redfish api access.
	Redfish time.Duration
	// SSH timeout for the ssh access.
	SSH time.Duration
}

// BMC is the holder struct for BMC access through redfish & ssh.
type BMC struct {
	host              string
	redfishUser       User
	sshUser           User
	sshPort           uint16
	systemIndex       int
	powerControlIndex int

	timeOuts TimeOuts

	sshSessionForSerialConsole *ssh.Session
}

// New returns a new BMC struct. The default system index to be used in redfish requests is 0.
// Use SetSystemIndex to modify it.
func New(host string, redfishUser, sshUser User, sshPort uint16, timeOuts TimeOuts) (*BMC, error) {
	glog.V(100).Infof("Initializing new BMC structure with the following params: %s, %v, %v, %v (system index = 0)",
		host, redfishUser, sshUser, timeOuts)

	errMsgs := []string{}
	if host == "" {
		errMsgs = append(errMsgs, "host is empty")
	}

	if redfishUser.Name == "" {
		errMsgs = append(errMsgs, "redfish user's name is empty")
	}

	if redfishUser.Password == "" {
		errMsgs = append(errMsgs, "redfish user's password is empty")
	}

	if sshUser.Name == "" {
		errMsgs = append(errMsgs, "ssh user's name is empty")
	}

	if sshPort == 0 {
		errMsgs = append(errMsgs, "ssh port is zero")
	}

	if sshUser.Password == "" {
		errMsgs = append(errMsgs, "ssh user's password is empty")
	}

	if timeOuts.Redfish == 0 {
		errMsgs = append(errMsgs, "redfish timeout is 0")
	}

	if timeOuts.SSH == 0 {
		errMsgs = append(errMsgs, "ssh timeout is 0")
	}

	// Build final error msg in case there were some error/s validating the input params.
	if len(errMsgs) > 0 {
		errMsg := ""

		for i, msg := range errMsgs {
			if i != 0 {
				errMsg += ", "
			}

			errMsg += msg
		}

		glog.V(100).Infof("Failed to initialize BMC: %s", errMsg)

		return nil, errors.New(errMsg)
	}

	return &BMC{
		host:              host,
		redfishUser:       redfishUser,
		sshUser:           sshUser,
		sshPort:           sshPort,
		timeOuts:          timeOuts,
		systemIndex:       0,
		powerControlIndex: 0,
	}, nil
}

// SetSystemIndex sets the system index to be used in redfish api requests.
func (bmc *BMC) SetSystemIndex(index int) error {
	glog.V(100).Infof("Setting default Redfish System Index to %d", index)

	if index < 0 {
		glog.V(100).Infof("Invalid system index %d: must be >= 0", index)

		return fmt.Errorf("invalid index %d", index)
	}

	bmc.systemIndex = index

	return nil
}

// SetPowerControlIndex sets the power control index to be used in Redfish API requests.
func (bmc *BMC) SetPowerControlIndex(index int) error {
	glog.V(100).Infof("Setting default Redfish Power Control Index to %d", index)

	if index < 0 {
		glog.V(100).Infof("Invalid power control index %d: must be >= 0", index)

		return fmt.Errorf("invalid index %d", index)
	}

	bmc.powerControlIndex = index

	return nil
}

// SystemManufacturer gets system's manufacturer from the BMC's RedFish API endpoint.
func (bmc *BMC) SystemManufacturer() (string, error) {
	glog.V(100).Infof("Getting SystemManufacturer param from bmc's redfish endpoint")

	redfishClient, cancel, err := redfishConnect(
		bmc.host,
		bmc.redfishUser.Name, bmc.redfishUser.Password,
		bmc.timeOuts.Redfish)
	if err != nil {
		glog.V(100).Infof("Redfish connection error: %v", err)

		return "", fmt.Errorf("redfish connection error: %w", err)
	}

	defer func() {
		redfishClient.Logout()
		cancel()
	}()

	system, err := redfishGetSystem(redfishClient, bmc.systemIndex)
	if err != nil {
		glog.V(100).Infof("Failed to get redfish system: %v", err)

		return "", fmt.Errorf("failed to get redfish system: %w", err)
	}

	return system.Manufacturer, nil
}

// IsSecureBootEnabled returns whether the SecureBoot feature is enabled using the BMC's RedFish API endpoint.
func (bmc *BMC) IsSecureBootEnabled() (bool, error) {
	glog.V(100).Infof("Getting secure boot status from bmc's redfish endpoint")

	redfishClient, cancel, err := redfishConnect(
		bmc.host,
		bmc.redfishUser.Name, bmc.redfishUser.Password,
		bmc.timeOuts.Redfish)
	if err != nil {
		glog.V(100).Infof("Redfish connection error: %v", err)

		return false, fmt.Errorf("redfish connection error: %w", err)
	}

	defer func() {
		redfishClient.Logout()
		cancel()
	}()

	sboot, err := redfishGetSystemSecureBoot(redfishClient, bmc.systemIndex)
	if err != nil {
		glog.V(100).Infof("Failed to get redfish system's secure boot: %v", err)

		return false, fmt.Errorf("failed to get secure boot: %w", err)
	}

	return sboot.SecureBootEnable, nil
}

// SecureBootEnable enables the SecureBoot feature using the BMC's RedFish API endpoint.
func (bmc *BMC) SecureBootEnable() error {
	glog.V(100).Infof("Enabling secure boot from bmc's redfish endpoint")

	redfishClient, cancel, err := redfishConnect(
		bmc.host,
		bmc.redfishUser.Name, bmc.redfishUser.Password,
		bmc.timeOuts.Redfish)
	if err != nil {
		glog.V(100).Infof("Redfish connection error: %v", err)

		return fmt.Errorf("redfish connection error: %w", err)
	}

	defer func() {
		redfishClient.Logout()
		cancel()
	}()

	sboot, err := redfishGetSystemSecureBoot(redfishClient, bmc.systemIndex)
	if err != nil {
		glog.V(100).Infof("Failed to get redfish system's secure boot: %v", err)

		return fmt.Errorf("failed to get secure boot: %w", err)
	}

	if sboot.SecureBootEnable {
		glog.V(100).Infof("Failed to enable secure boot: it is already enabled")

		return fmt.Errorf("secure boot is already enabled")
	}

	sboot.SecureBootEnable = true

	err = sboot.Update()
	if err != nil {
		glog.V(100).Infof("Failed to enable secure boot: %v", err)

		return fmt.Errorf("failed to enable secure boot: %w", err)
	}

	return nil
}

// SecureBootDisable disables the SecureBoot feature using the BMC's RedFish API endpoint.
func (bmc *BMC) SecureBootDisable() error {
	glog.V(100).Infof("Disabling secure boot from bmc's redfish endpoint")

	redfishClient, cancel, err := redfishConnect(
		bmc.host,
		bmc.redfishUser.Name, bmc.redfishUser.Password,
		bmc.timeOuts.Redfish)
	if err != nil {
		glog.V(100).Infof("Redfish connection error: %v", err)

		return fmt.Errorf("redfish connection error: %w", err)
	}

	defer func() {
		redfishClient.Logout()
		cancel()
	}()

	sboot, err := redfishGetSystemSecureBoot(redfishClient, bmc.systemIndex)
	if err != nil {
		glog.V(100).Infof("Failed to get redfish system's secure boot: %v", err)

		return fmt.Errorf("failed to get secure boot: %w", err)
	}

	if !sboot.SecureBootEnable {
		glog.V(100).Infof("Failed to disable secure boot: it is already disabled")

		return fmt.Errorf("secure boot is already disabled")
	}

	sboot.SecureBootEnable = false

	err = sboot.Update()
	if err != nil {
		glog.V(100).Infof("Failed to disable secure boot: %v", err)

		return fmt.Errorf("failed to disable secure boot: %w", err)
	}

	return nil
}

// SystemResetAction performs the specified reset action against the system.
func (bmc *BMC) SystemResetAction(action redfish.ResetType) error {
	glog.V(100).Infof("Performing reset action %s from the bmc's redfish endpoint")

	redfishClient, cancel, err := redfishConnect(
		bmc.host,
		bmc.redfishUser.Name, bmc.redfishUser.Password,
		bmc.timeOuts.Redfish)
	if err != nil {
		glog.V(100).Infof("Redfish connection error: %v", err)

		return fmt.Errorf("redfish connection error: %w", err)
	}

	defer func() {
		redfishClient.Logout()
		cancel()
	}()

	system, err := redfishGetSystem(redfishClient, bmc.systemIndex)
	if err != nil {
		glog.V(100).Infof("Failed to get redfish system: %v", err)

		return fmt.Errorf("failed to get redfish system: %w", err)
	}

	return system.Reset(action)
}

// SystemForceReset performs a (non-graceful) forced system reset using redfish API.
func (bmc *BMC) SystemForceReset() error {
	return bmc.SystemResetAction(redfish.ForceRestartResetType)
}

// SystemGracefulShutdown performs a graceful shutdown using the redfish API.
func (bmc *BMC) SystemGracefulShutdown() error {
	return bmc.SystemResetAction(redfish.GracefulShutdownResetType)
}

// SystemPowerOn powers on the system using the redfish API.
func (bmc *BMC) SystemPowerOn() error {
	return bmc.SystemResetAction(redfish.OnResetType)
}

// SystemPowerCycle power cycles the system using the redfish API.
func (bmc *BMC) SystemPowerCycle() error {
	return bmc.SystemResetAction(redfish.PowerCycleResetType)
}

// PowerUsage returns the current power usage of the chassis in watts using the redfish API. This method uses the first
// chassis with a power link and the power control index for the BMC client.
func (bmc *BMC) PowerUsage() (float32, error) {
	glog.V(100).Info("Collecting current power usage from bmc's redfish endpoint")

	redfishClient, cancel, err := redfishConnect(
		bmc.host,
		bmc.redfishUser.Name, bmc.redfishUser.Password,
		bmc.timeOuts.Redfish)
	if err != nil {
		glog.V(100).Infof("Redfish connection error: %v", err)

		return 0.0, fmt.Errorf("redfish connection error: %w", err)
	}

	defer func() {
		redfishClient.Logout()
		cancel()
	}()

	powerControl, err := redfishGetPowerControl(redfishClient, bmc.powerControlIndex)
	if err != nil {
		glog.V(100).Infof("Failed to get redfish power control: %v", err)

		return 0.0, fmt.Errorf("failed to get redfish power control: %w", err)
	}

	return powerControl.PowerConsumedWatts, nil
}

// CreateCLISSHSession creates a ssh Session to the host.
func (bmc *BMC) CreateCLISSHSession() (*ssh.Session, error) {
	glog.V(100).Infof("Creating SSH session to run commands in the BMC's CLI.")

	config := &ssh.ClientConfig{
		User: bmc.sshUser.Name,
		Auth: []ssh.AuthMethod{
			ssh.Password(bmc.sshUser.Password),
			ssh.KeyboardInteractive(func(user, instruction string, questions []string,
				echos []bool) (answers []string, err error) {
				answers = make([]string, len(questions))
				// The second parameter is unused
				for n := range questions {
					answers[n] = bmc.sshUser.Password
				}

				return answers, nil
			}),
		},
		Timeout:         bmc.timeOuts.SSH,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Establish SSH connection
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", bmc.host, bmc.sshPort), config)
	if err != nil {
		glog.V(100).Infof("Failed to connect to BMC's SSH server: %v", err)

		return nil, fmt.Errorf("failed to connect to BMC's SSH server: %w", err)
	}

	// Create a session
	session, err := client.NewSession()
	if err != nil {
		glog.V(100).Infof("Failed to create a new SSH session: %v", err)

		return nil, fmt.Errorf("failed to create a new ssh session: %w", err)
	}

	return session, nil
}

// RunCLICommand runs a CLI command in the BMC's console. This method will block until the command
// has finished, and its output is copied to stdout and/or stderr if applicable. If combineOutput is true,
// stderr content is merged in stdout. The timeout param is used to avoid the caller to be stuck forever
// in case something goes wrong or the command is stuck.
func (bmc *BMC) RunCLICommand(cmd string, combineOutput bool, timeout time.Duration) (
	stdout string, stderr string, err error) {
	glog.V(100).Infof("Running CLI command in BMC's CLI: %s", cmd)

	sshSession, err := bmc.CreateCLISSHSession()
	if err != nil {
		glog.V(100).Infof("Failed to connect to CLI: %v", err)

		return "", "", fmt.Errorf("failed to connect to CLI: %w", err)
	}

	defer sshSession.Close()

	var stdoutBuffer, stderrBuffer bytes.Buffer
	if !combineOutput {
		sshSession.Stdout = &stdoutBuffer
		sshSession.Stderr = &stderrBuffer
	}

	var combinedOutput []byte

	errCh := make(chan error)
	go func() {
		var err error
		if combineOutput {
			combinedOutput, err = sshSession.CombinedOutput(cmd)
		} else {
			err = sshSession.Run(cmd)
		}
		errCh <- err
	}()

	timeoutCh := time.After(timeout)

	select {
	case <-timeoutCh:
		glog.V(100).Info("CLI command timeout")

		return stdoutBuffer.String(), stderrBuffer.String(), fmt.Errorf("timeout running command")
	case err := <-errCh:
		glog.V(100).Info("Command run error: %v", err)

		if err != nil {
			return stdoutBuffer.String(), stderrBuffer.String(), fmt.Errorf("command run error: %w", err)
		}
	}

	if combineOutput {
		return string(combinedOutput), "", nil
	}

	return stdoutBuffer.String(), stderrBuffer.String(), nil
}

// OpenSerialConsole opens the serial console port. The console is tunneled in an underlying (CLI)
// ssh session that is opened in the BMC's ssh server. If openConsoleCliCmd is
// provided, it will be sent to the BMC's cli. Otherwise, a best effort will
// be made to run the appropriate cli command based on the system manufacturer.
func (bmc *BMC) OpenSerialConsole(openConsoleCliCmd string) (io.Reader, io.WriteCloser, error) {
	glog.V(100).Infof("Opening serial console on %v.", bmc.host)

	if bmc.sshSessionForSerialConsole != nil {
		glog.V(100).Infof("There is already a serial console opened for %v's BMC. Use OpenSerialConsole() first.",
			bmc.host)

		return nil, nil, fmt.Errorf("there is already a serial console opened for %v's BMC", bmc.host)
	}

	cliCmd := openConsoleCliCmd
	if cliCmd == "" {
		// no cli command to get console port was provided, try to guess based on
		// manufacturer.
		manufacturer, err := bmc.SystemManufacturer()
		if err != nil {
			glog.V(100).Infof("Failed to get redifsh system manufacturer for %v: %v", bmc.host, err)

			return nil, nil, fmt.Errorf("failed to get redfish system manufacturer for %v: %w", bmc.host, err)
		}

		var found bool
		if cliCmd, found = cliCmdSerialConsole[manufacturer]; !found {
			glog.V(100).Infof("CLI command to get serial console not found for manufacturer for %v: %v",
				bmc.host, manufacturer)

			return nil, nil, fmt.Errorf("cli command to get serial console not found for manufacturer for %v: %v",
				bmc.host, manufacturer)
		}
	}

	sshSession, err := bmc.CreateCLISSHSession()
	if err != nil {
		glog.V(100).Infof("Failed to create underlying ssh session for %v: %v", bmc.host, err)

		return nil, nil, fmt.Errorf("failed to create underlying ssh session for %v: %w", bmc.host, err)
	}

	// Pipes need to be retrieved before session.Start()
	reader, err := sshSession.StdoutPipe()
	if err != nil {
		glog.V(100).Infof("Failed to get stdout pipe from %v's ssh session: %v", bmc.host, err)

		_ = sshSession.Close()

		return nil, nil, fmt.Errorf("failed to get stdout pipe from %v's ssh session: %w", bmc.host, err)
	}

	writer, err := sshSession.StdinPipe()
	if err != nil {
		glog.V(100).Infof("Failed to get stdin pipe from from %v's ssh session: %w", bmc.host, err)

		_ = sshSession.Close()

		return nil, nil, fmt.Errorf("failed to get stdin pipe from %v's ssh session: %w", bmc.host, err)
	}

	err = sshSession.Start(cliCmd)
	if err != nil {
		glog.V(100).Infof("Failed to start CLI command %q on %v: %v", cliCmd, bmc.host, err)

		_ = sshSession.Close()

		return nil, nil, fmt.Errorf("failed to start serial console with cli command %q on %v: %w", cliCmd, bmc.host, err)
	}

	go func() { _ = sshSession.Wait() }()

	bmc.sshSessionForSerialConsole = sshSession

	return reader, writer, nil
}

// CloseSerialConsole closes the serial console's underlying ssh session.
func (bmc *BMC) CloseSerialConsole() error {
	glog.V(100).Infof("Closing serial console for %v.", bmc.host)

	if bmc.sshSessionForSerialConsole == nil {
		glog.V(100).Infof("No underlying ssh session found for %v. Please use OpenSerialConsole() first.", bmc.host)

		return fmt.Errorf("no underlying ssh session found for %v", bmc.host)
	}

	err := bmc.sshSessionForSerialConsole.Close()
	if err != nil {
		glog.V(100).Infof("Failed to close underlying ssh session for %v: %v", bmc.host, err)

		return fmt.Errorf("failed to close underlying ssh session for %v: %w", bmc.host, err)
	}

	bmc.sshSessionForSerialConsole = nil

	return nil
}

func redfishConnect(host, user, password string, sessionTimeout time.Duration) (
	*gofish.APIClient, context.CancelFunc, error) {
	gofishConfig := gofish.ClientConfig{
		Endpoint: "https://" + host,
		Username: user,
		Password: password,
		Insecure: true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), sessionTimeout)

	client, err := gofish.ConnectContext(ctx, gofishConfig)
	if err != nil {
		cancel()

		return nil, nil, fmt.Errorf("failed to connect to redfish endpoint: %w", err)
	}

	return client, cancel, nil
}

func redfishGetSystem(redfishClient *gofish.APIClient, index int) (*redfish.ComputerSystem, error) {
	systems, err := redfishClient.GetService().Systems()
	if err != nil {
		return nil, fmt.Errorf("failed to get systems: %w", err)
	}

	if len(systems) < index+1 {
		return nil, fmt.Errorf("invalid system index %d (base-index=0, num systems=%d)", index, len(systems))
	}

	return systems[index], nil
}

func redfishGetSystemSecureBoot(redfishClient *gofish.APIClient, systemIndex int) (*redfish.SecureBoot, error) {
	system, err := redfishGetSystem(redfishClient, systemIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to get redfish system: %w", err)
	}

	sboot, err := system.SecureBoot()
	if err != nil {
		return nil, err
	}

	return sboot, nil
}

// redfishGetPowerControl gets the specified PowerControl from the first chassis with a power link from the redfish API.
func redfishGetPowerControl(
	redfishClient *gofish.APIClient, powerControlIndex int) (*redfish.PowerControl, error) {
	chassisCollection, err := redfishClient.GetService().Chassis()
	if err != nil {
		return nil, fmt.Errorf("failed to get chassis collection: %w", err)
	}

	for chassisIndex, chassis := range chassisCollection {
		power, err := chassis.Power()
		if err != nil {
			return nil, fmt.Errorf("failed to get power for chassis index %d: %w", chassisIndex, err)
		}

		if power == nil {
			continue
		}

		if powerControlIndex >= len(power.PowerControl) {
			return nil, fmt.Errorf(
				"invalid power control index %d (base-index=0, num power control=%d)", powerControlIndex, len(power.PowerControl))
		}

		return &power.PowerControl[powerControlIndex], nil
	}

	return nil, fmt.Errorf("failed to get power control: no chassis with power link found")
}
