export enum Unit {
  Kilo = 'kilo',
  Gram = 'g',
  Milligram = 'mg',
  Liter = 'l',
  Milliliter = 'ml',
  Teaspoon = 'tsp',
  Tablespoon = 'tbsp',
  Cup = 'cup',
  Quart = 'qt',
  Countable = 'countable',
  Uncountable = 'uncountable',
}

export const UnitOptions = [
  { value: Unit.Kilo, label: 'kg' },
  { value: Unit.Gram, label: 'g' },
  { value: Unit.Milligram, label: 'mg' },
  { value: Unit.Liter, label: 'l' },
  { value: Unit.Milliliter, label: 'ml' },
  { value: Unit.Teaspoon, label: 'tsp' },
  { value: Unit.Tablespoon, label: 'tbsp' },
  { value: Unit.Cup, label: 'cup' },
  { value: Unit.Quart, label: 'qt' },
  { value: Unit.Countable, label: 'unit(s)' },
  { value: Unit.Uncountable, label: 'some' },
];

export const DefaultUnit = Unit.Uncountable;
