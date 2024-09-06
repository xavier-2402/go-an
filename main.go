package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"net/http"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

func listVm(projectID string, zone string) {
	credentialsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credentialsPath == "" {
		fmt.Println("La variable de entorno GOOGLE_APPLICATION_CREDENTIALS no está establecida.")
		return
	}
	ctx := context.Background()
	computeService, err := compute.NewService(ctx, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		log.Fatalf("Failed to create compute service: %v", err)
	}

	instanceList, err := computeService.Instances.List(projectID, zone).Context(ctx).Do()
	if err != nil {
		log.Fatalf("Failed to list instances: %v", err)
	}

	for _, instance := range instanceList.Items {
		fmt.Printf("Instance Name: %s with status %s\n", instance.Name, instance.Status)
	}
}

func getVm(projectID string, zone string, instanceId string) {
	credentialsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credentialsPath == "" {
		fmt.Println("La variable de entorno GOOGLE_APPLICATION_CREDENTIALS no está establecida.")
		return
	}
	ctx := context.Background()

	computeService, err := compute.NewService(ctx, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		log.Fatalf("Failed to create compute service: %v", err)
	}

	// Intentar obtener la instancia
	instance, err := computeService.Instances.Get(projectID, zone, instanceId).Context(ctx).Do()
	if err != nil {
		log.Fatalf("Failed to get instance: %v", err)
		return
	}

	// Imprimir el estado de la instancia
	fmt.Printf("Instance %s retrieved successfully. Status: %s\n", instanceId, instance.Status)
}

func startVM(projectID string, zone string, instanceId string) {
	credentialsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credentialsPath == "" {
		fmt.Println("La variable de entorno GOOGLE_APPLICATION_CREDENTIALS no está establecida.")
		return
	}
	ctx := context.Background()

	computeService, err := compute.NewService(ctx, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		log.Fatalf("Failed to create compute service: %v", err)
		return
	}

	// Intentar obtener la instancia
	instance, err := computeService.Instances.Get(projectID, zone, instanceId).Context(ctx).Do()

	if err != nil {
		log.Fatalf("Failed to get instance: %v", err)
		return
	}
	if instance.Status != "TERMINATED" {
		log.Fatalf("instance is not in TERMINATED state")
	}

	// Iniciar la instancia
	op, err := computeService.Instances.Start(projectID, zone, instanceId).Context(ctx).Do()
	if err != nil {
		log.Fatalf("failed to start instance: %v", err)
	}

	err = wait(ctx, computeService, projectID, zone, op.Name)
	if err != nil {
		log.Fatalf("wait for operation failed: %v", err)
	}

	fmt.Printf("Instance %s started successfully.\n", instanceId)
}

func suspendVM(projectID string, zone string, instanceId string) {
	credentialsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credentialsPath == "" {
		fmt.Println("La variable de entorno GOOGLE_APPLICATION_CREDENTIALS no está establecida.")
		return
	}
	ctx := context.Background()

	computeService, err := compute.NewService(ctx, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		log.Fatalf("Failed to create compute service: %v", err)
		return
	}

	// Intentar obtener la instancia
	instance, err := computeService.Instances.Get(projectID, zone, instanceId).Context(ctx).Do()

	if err != nil {
		log.Fatalf("Failed to get instance: %v", err)
		return
	}
	if instance.Status == "SUSPENDED" {
		log.Fatalf("instance is already in SUSPENDED state")
	}

	// Suspender la instancia
	op, err := computeService.Instances.Suspend(projectID, zone, instanceId).Context(ctx).Do()
	if err != nil {
		log.Fatalf("failed to suspend instance: %v", err)
	}

	err = wait(ctx, computeService, projectID, zone, op.Name)
	if err != nil {
		log.Fatalf("wait for operation failed: %v", err)
	}

	fmt.Printf("Instance %s suspended successfully.\n", instanceId)
}

func stopVM(projectID string, zone string, instanceId string) {
	credentialsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credentialsPath == "" {
		fmt.Println("La variable de entorno GOOGLE_APPLICATION_CREDENTIALS no está establecida.")
		return
	}
	ctx := context.Background()

	computeService, err := compute.NewService(ctx, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		log.Fatalf("Failed to create compute service: %v", err)
		return
	}

	// Intentar obtener la instancia
	instance, err := computeService.Instances.Get(projectID, zone, instanceId).Context(ctx).Do()

	if err != nil {
		log.Fatalf("Failed to get instance: %v", err)
		return
	}
	if instance.Status == "TERMINATED" {
		log.Fatalf("instance is already in TERMINATED state")
	}

	// Suspender la instancia
	op, err := computeService.Instances.Stop(projectID, zone, instanceId).Context(ctx).Do()
	if err != nil {
		log.Fatalf("failed to stop instance: %v", err)
	}

	err = wait(ctx, computeService, projectID, zone, op.Name)
	if err != nil {
		log.Fatalf("wait for operation failed: %v", err)
	}

	fmt.Printf("Instance %s stop successfully.\n", instanceId)
}

func wait(ctx context.Context, computeService *compute.Service, projectID, zone, operationName string) error {
	for {
		// Obtener el estado de la operación
		op, err := computeService.ZoneOperations.Get(projectID, zone, operationName).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("could not get operation status: %w", err)
		}

		// Verificar si la operación ha finalizado
		if op.Status == "DONE" {
			if op.Error != nil && len(op.Error.Errors) > 0 {
				// La operación terminó con un error
				return fmt.Errorf("operation completed with error: %v", op.Error.Errors)
			}
			// La operación terminó correctamente
			return nil
		}

		// Dormir por un tiempo antes de volver a consultar el estado
		time.Sleep(2 * time.Second)
	}
}

func HelloWorld(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "¡Hola desde Google Cloud Functions en Go!")
}

func main() {
	projectID := "steel-totality-430001-n0"
	zone := "us-central1-c"
	instanceId := "2557316538645784744"

	// Listar instancias
	//listVm(projectID, zone)

	// Obtener una instancia de máquina virtual
	getVm(projectID, zone, instanceId)

	//stopVM(projectID, zone, instanceId)
	startVM(projectID, zone, instanceId)

	//suspendVM(projectID, zone, instanceId)
	getVm(projectID, zone, instanceId)

}
