package hash

import (
	"fmt"
	"hash"
	"slices"

	"golang.org/x/exp/maps"
	v1 "k8s.io/api/core/v1"
)

func Secret(h hash.Hash, s *v1.Secret) error {
	if _, err := h.Write([]byte(s.GetName())); err != nil {
		return fmt.Errorf("could not write to hash: %w", err)
	}
	return binaryData(h, s.Data)
}

func ConfigMap(h hash.Hash, cm *v1.ConfigMap) error {
	if _, err := h.Write([]byte(cm.GetName())); err != nil {
		return fmt.Errorf("could not write to hash: %w", err)
	}
	if err := stringData(h, cm.Data); err != nil {
		return err
	}
	return binaryData(h, cm.BinaryData)
}

func SecretKeys(h hash.Hash, referencedKeys []string, s *v1.Secret) error {
	if _, err := h.Write([]byte(s.GetName())); err != nil {
		return fmt.Errorf("could not write to hash: %w", err)
	}
	return binaryDataKeys(h, referencedKeys, s.Data)
}

func ConfigMapKeys(h hash.Hash, referencedKeys []string, cm *v1.ConfigMap) error {
	if _, err := h.Write([]byte(cm.GetName())); err != nil {
		return fmt.Errorf("could not write to hash: %w", err)
	}
	if err := stringDataKeys(h, referencedKeys, cm.Data); err != nil {
		return err
	}
	return binaryDataKeys(h, referencedKeys, cm.BinaryData)
}

func binaryDataKeys(h hash.Hash, referencedKeys []string, data map[string][]byte) error {
	ks := slices.Clone(referencedKeys)
	slices.Sort(ks)
	for _, k := range ks {
		if v, ok := data[k]; ok {
			if _, err := h.Write([]byte(k)); err != nil {
				return fmt.Errorf("could not write to hash: %w", err)
			}
			if _, err := h.Write(v); err != nil {
				return fmt.Errorf("could not write to hash: %w", err)
			}
		}
	}
	return nil
}

func stringDataKeys(h hash.Hash, referencedKeys []string, data map[string]string) error {
	ks := slices.Clone(referencedKeys)
	slices.Sort(ks)
	for _, k := range ks {
		if v, ok := data[k]; ok {
			if _, err := h.Write([]byte(k)); err != nil {
				return fmt.Errorf("could not write to hash: %w", err)
			}
			if _, err := h.Write([]byte(v)); err != nil {
				return fmt.Errorf("could not write to hash: %w", err)
			}
		}
	}
	return nil
}

func binaryData(h hash.Hash, data map[string][]byte) error {
	ks := maps.Keys(data)
	slices.Sort(ks)
	for _, k := range ks {
		v := data[k]
		if _, err := h.Write([]byte(k)); err != nil {
			return fmt.Errorf("could not write to hash: %w", err)
		}
		if _, err := h.Write(v); err != nil {
			return fmt.Errorf("could not write to hash: %w", err)
		}
	}
	return nil
}

func stringData(h hash.Hash, data map[string]string) error {
	ks := maps.Keys(data)
	slices.Sort(ks)
	for _, k := range ks {
		v := data[k]
		if _, err := h.Write([]byte(k)); err != nil {
			return fmt.Errorf("could not write to hash: %w", err)
		}
		if _, err := h.Write([]byte(v)); err != nil {
			return fmt.Errorf("could not write to hash: %w", err)
		}
	}
	return nil
}
