import {
  Control,
  ControllerProps,
  FieldPath,
  FieldValues,
} from 'react-hook-form';

import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '../ui/form';
import { MultiSelectOption } from '../ui/multiselector';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '../ui/select';

interface FormElementProps<
  TFieldValues extends FieldValues = FieldValues,
  TName extends FieldPath<TFieldValues> = FieldPath<TFieldValues>,
> {
  control: Control<TFieldValues>;
  fieldName: TName;
  label?: string;
  render: ControllerProps<TFieldValues, TName>['render'];
}

export const FormElement = <
  TFieldValues extends FieldValues = FieldValues,
  TName extends FieldPath<TFieldValues> = FieldPath<TFieldValues>,
>({
  control,
  fieldName,
  label,
  render,
}: FormElementProps<TFieldValues, TName>) => (
  <FormField
    control={control}
    name={fieldName}
    render={(renderProps) => (
      <FormItem>
        {label && <FormLabel>{label}</FormLabel>}
        <FormControl>{render(renderProps)}</FormControl>
        <FormMessage />
      </FormItem>
    )}
  />
);

export const FormSelectElement = <
  TFieldValues extends FieldValues = FieldValues,
  TName extends FieldPath<TFieldValues> = FieldPath<TFieldValues>,
>({
  control,
  fieldName,
  label,
  placeholder,
  options,
  disabled,
}: Omit<FormElementProps<TFieldValues, TName>, 'render'> & {
  placeholder?: string;
  options: MultiSelectOption[];
  disabled?: boolean;
}) => (
  <FormField
    control={control}
    name={fieldName}
    render={(renderProps) => (
      <FormItem>
        {label && <FormLabel>{label}</FormLabel>}
        <Select
          onValueChange={renderProps.field.onChange}
          defaultValue={renderProps.field.value}
        >
          <FormControl>
            <SelectTrigger disabled={disabled}>
              <SelectValue placeholder={placeholder} />
            </SelectTrigger>
          </FormControl>
          <SelectContent>
            {options.map((option) => (
              <SelectItem key={option.value} value={option.value}>
                {option.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </FormItem>
    )}
  />
);
