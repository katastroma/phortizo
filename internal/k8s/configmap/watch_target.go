//revive:disable:package-comments
package configmap

import (
	"context"
	"fmt"

	"github.com/katastroma/phortizo/internal/source"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	// TypeLabel is the label key used to identify ConfigMap types.
	TypeLabel = "katastroma.org/type"

	// WatchTargetType is the label value for watch target ConfigMaps.
	WatchTargetType = "watch-target"

	// CredentialSecretAnnotation is the annotation key pointing to the
	// credential Secret name in the same namespace.
	CredentialSecretAnnotation = "katastroma.org/credential-secret"
)

// GetWatchTarget reads a single watch target ConfigMap by name.
func GetWatchTarget(ctx context.Context, client kubernetes.Interface, namespace, name string) (source.WatchTarget, error) {
	cm, err := client.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return source.WatchTarget{}, fmt.Errorf("reading configmap %s/%s: %w", namespace, name, err)
	}

	return watchTargetFromConfigMap(cm)
}

// ListWatchTargets lists all watch target ConfigMaps in the given namespace
// and returns the deserialized watch targets.
func ListWatchTargets(ctx context.Context, client kubernetes.Interface, namespace string) ([]source.WatchTarget, error) {
	selector := fmt.Sprintf("%s=%s", TypeLabel, WatchTargetType)
	list, err := client.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, fmt.Errorf("listing watch target configmaps in %s: %w", namespace, err)
	}

	targets := make([]source.WatchTarget, 0, len(list.Items))
	for _, cm := range list.Items {
		target, err := watchTargetFromConfigMap(&cm)
		if err != nil {
			return nil, fmt.Errorf("deserializing configmap %s/%s: %w", namespace, cm.Name, err)
		}

		targets = append(targets, target)
	}

	return targets, nil
}

func watchTargetFromConfigMap(cm *corev1.ConfigMap) (source.WatchTarget, error) {
	target, err := source.WatchTargetFromConfigMap(cm.Data)
	if err != nil {
		return source.WatchTarget{}, err
	}

	target.Name = cm.Name
	target.CredentialSecret = cm.Annotations[CredentialSecretAnnotation]
	return target, nil
}

// PutWatchTarget creates or updates a watch target ConfigMap in the given
// namespace. If the target has a CredentialSecret, it is set as an annotation.
func PutWatchTarget(
	ctx context.Context,
	client kubernetes.Interface,
	namespace, name string,
	target source.WatchTarget,
) error {
	configmaps := client.CoreV1().ConfigMaps(namespace)
	data := target.MarshalConfigMap()
	annotations := buildAnnotations(target.CredentialSecret)
	labels := map[string]string{TypeLabel: WatchTargetType}

	existing, err := configmaps.Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		existing.Data = data
		existing.Labels = labels
		existing.Annotations = annotations
		_, err = configmaps.Update(ctx, existing, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("updating configmap %s/%s: %w", namespace, name, err)
		}

		return nil
	}

	_, err = configmaps.Create(ctx, &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Data: data,
	}, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("creating configmap %s/%s: %w", namespace, name, err)
	}

	return nil
}

func buildAnnotations(credentialSecret string) map[string]string {
	if credentialSecret == "" {
		return nil
	}

	return map[string]string{CredentialSecretAnnotation: credentialSecret}
}
