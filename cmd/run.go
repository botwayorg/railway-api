package cmd

import (
	"context"
	goErr "errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/botwayorg/railway-api/entity"
	"github.com/botwayorg/railway-api/errors"
	"github.com/botwayorg/railway-api/ui"
)

var RAIL_PORT = 4411

func (h *Handler) getEnvironment(ctx context.Context, environmentName string) (*entity.Environment, error) {
	if environmentName == "" {
		return h.ctrl.GetCurrentEnvironment(ctx)
	}

	return h.ctrl.GetEnvironmentByName(ctx, environmentName)
}

func (h *Handler) Run(ctx context.Context, req *entity.CommandRequest) error {
	isEphemeral := false

	for _, arg := range req.Args {
		if arg == "--ephemeral" {
			isEphemeral = true
		}
	}

	parsedArgs := make([]string, 0)

	rgxEnvironment, err := regexp.Compile("--environment=(.*)")
	if err != nil {
		return err
	}

	rgxService, err := regexp.Compile("--service=(.*)")
	if err != nil {
		return err
	}

	targetEnvironment := ""
	var targetServiceName *string

	// Parse --environment={ENV} and --service={SERVICE} from args
	for _, arg := range req.Args {
		if matched := rgxEnvironment.FindStringSubmatch(arg); matched != nil {
			if len(matched) < 2 {
				return goErr.New("missing environment selection! \n(e.g --environment=production)")
			}

			targetEnvironment = matched[1]
		} else if matched := rgxService.FindStringSubmatch(arg); matched != nil {
			if len(matched) < 2 {
				return goErr.New("missing service selection! \n(e.g --service=serviceName)")
			}

			targetServiceName = &matched[1]
		} else {
			parsedArgs = append(parsedArgs, arg)
		}
	}

	projectCfg, err := h.ctrl.GetProjectConfigs(ctx)
	if err != nil {
		return err
	}

	environment, err := h.getEnvironment(ctx, targetEnvironment)
	if err != nil {
		return err
	}
	// Add something to the ephemeral env name
	if isEphemeral {
		environmentName := fmt.Sprintf("%s-ephemeral", environment.Name)
		fmt.Printf("Spinning up Ephemeral Environment: %s\n", ui.BlueText(environmentName))
		// Create new environment for this run

		environment, err = h.ctrl.CreateEphemeralEnvironment(ctx, &entity.CreateEphemeralEnvironmentRequest{
			Name:              environmentName,
			ProjectID:         projectCfg.Project,
			BaseEnvironmentID: environment.Id,
		})

		if err != nil {
			return err
		}

		fmt.Println("Done!")
	}

	envs, err := h.ctrl.GetEnvs(ctx, environment, targetServiceName)

	if err != nil {
		return err
	}

	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	hasDockerfile := true

	if _, err := os.Stat(fmt.Sprintf("%s/Dockerfile", pwd)); os.IsNotExist(err) {
		hasDockerfile = false
	}

	if len(parsedArgs) == 0 && hasDockerfile {
		return h.runInDocker(ctx, pwd, envs)
	} else if len(parsedArgs) == 0 {
		return errors.CommandNotSpecified
	}

	cmd := exec.CommandContext(ctx, parsedArgs[0], parsedArgs[1:]...)
	cmd.Env = os.Environ()

	// Inject railway envs
	for k, v := range *envs {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%+v", k, v))
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	cmd.Stdin = os.Stdin
	catchSignals(ctx, cmd, nil)

	err = cmd.Run()

	if isEphemeral {
		// Teardown Environment
		fmt.Println("Tearing down ephemeral environment...")
		err := h.ctrl.DeleteEnvironment(ctx, &entity.DeleteEnvironmentRequest{
			EnvironmentId: environment.Id,
			ProjectID:     projectCfg.Project,
		})

		if err != nil {
			return err
		}

		fmt.Println("Done!")
	}

	if err != nil {
		fmt.Println(err.Error())

		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}

		os.Exit(1)
	}

	printLooksGood()

	return nil
}

func (h *Handler) runInDocker(ctx context.Context, pwd string, envs *entity.Envs) error {
	// Start building the image
	projectCfg, err := h.ctrl.GetProjectConfigs(ctx)
	if err != nil {
		return err
	}

	project, err := h.ctrl.GetProject(ctx, projectCfg.Project)
	if err != nil {
		return err
	}

	// Strip characters not allowed in Docker image names
	environment, err := h.ctrl.GetCurrentEnvironment(ctx)
	if err != nil {
		return err
	}

	sanitiser := regexp.MustCompile(`[^A-Za-z0-9_-]`)
	imageNameWithoutNsOrTag := strings.ToLower(sanitiser.ReplaceAllString(project.Name, "") + "-" + sanitiser.ReplaceAllString(environment.Name, ""))
	image := fmt.Sprintf("railway-local/%s:latest", imageNameWithoutNsOrTag)

	buildArgs := []string{"build", "-q", "-t", image, pwd}

	// Build up env
	for k, v := range *envs {
		buildArgs = append(buildArgs, "--build-arg", fmt.Sprintf("%s=\"%+v\"", k, v))
	}

	buildCmd := exec.CommandContext(ctx, "docker", buildArgs...)
	fmt.Printf("Building %s from Dockerfile...\n", ui.GreenText(image))

	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr

	err = buildCmd.Start()

	if err != nil {
		return err
	}

	err = buildCmd.Wait()

	if err != nil {
		return err
	}

	fmt.Printf("🎉 Built %s\n", ui.GreenText(image))

	// Attempt to use
	internalPort := envs.Get("PORT")

	externalPort, err := getAvailablePort()
	if err != nil {
		return err
	}

	if internalPort == "" {
		internalPort = externalPort
	}

	// Start running the image
	fmt.Printf("🚂 Running at %s\n\n", ui.GreenText(fmt.Sprintf("127.0.0.1:%s", externalPort)))

	runArgs := []string{"run", "--init", "--rm", "-p", fmt.Sprintf("127.0.0.1:%s:%s", externalPort, internalPort), "-e", fmt.Sprintf("PORT=%s", internalPort), "-d"}
	// Build up env
	for k, v := range *envs {
		runArgs = append(runArgs, "-e", fmt.Sprintf("%s=%+v", k, v))
	}

	runArgs = append(runArgs, image)

	// Run the container
	rawContainerId, err := exec.CommandContext(ctx, "docker", runArgs...).Output()
	if err != nil {
		return err
	}

	// Get the container ID
	containerId := strings.TrimSpace(string(rawContainerId))

	// Attach to the container
	logCmd := exec.CommandContext(ctx, "docker", "logs", "-f", containerId)
	logCmd.Stdout = os.Stdout
	logCmd.Stderr = os.Stderr

	err = logCmd.Start()

	if err != nil {
		return err
	}
	// Listen for cancel to remove the container
	catchSignals(ctx, logCmd, func() {
		err = exec.Command("docker", "rm", "-f", string(containerId)).Run()
	})

	err = logCmd.Wait()

	if err != nil && !strings.Contains(err.Error(), "255") {
		// 255 is a graceeful exit with ctrl + c
		return err
	}

	printLooksGood()

	return nil
}

func getAvailablePort() (string, error) {
	searchRange := 64

	for i := RAIL_PORT; i < RAIL_PORT+searchRange; i++ {
		if isAvailable(i) {
			return strconv.Itoa(i), nil
		}
	}

	return "", fmt.Errorf("couldn't find available port between %d and %d", RAIL_PORT, RAIL_PORT+searchRange)
}

func isAvailable(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		return false
	}

	_ = ln.Close()

	return true
}

func catchSignals(_ context.Context, cmd *exec.Cmd, onSignal context.CancelFunc) {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		err := cmd.Process.Signal(sig)

		if onSignal != nil {
			onSignal()
		}

		if err != nil {
			fmt.Println("Child process error: \n", err)
		}
	}()
}

func printLooksGood() {
	// Get space between last output and this message
	fmt.Println()
	fmt.Printf(
		"🚄 Looks good? Then put it on the train and deploy with `%s`!\n",
		ui.GreenText("railway up"),
	)
}
