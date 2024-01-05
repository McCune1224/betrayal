package inventory

func (ih *InventoryHandler) SetRole(roleName string) (string, error) {
	role, err := ih.m.Roles.GetByFuzzy(roleName)
	if err != nil {
		return "", err
	}

	ih.i.RoleName = role.Name
	err = ih.m.Inventories.UpdateRoleName(ih.i)
	if err != nil {
		return "", err
	}

	return role.Name, nil
}
