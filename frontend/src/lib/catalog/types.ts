export type CatalogKind = 'roles' | 'items' | 'abilities' | 'statuses' | 'perks' | 'categories';

export type CatalogRecord = {
  id: number;
  name: string;
  description: string;
  alignment?: string;
  rarity?: string;
  cost?: number;
  default_charges?: number;
  any_ability?: boolean;
  hour_duration?: number;
  abilities?: Array<{ id: number; name: string }>;
  perks?: Array<{ id: number; name: string }>;
};
