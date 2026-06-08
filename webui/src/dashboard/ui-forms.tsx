import { Controller, useForm } from 'react-hook-form';
import type { Control, FieldValues, Path, SubmitHandler } from 'react-hook-form';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { cn } from '@/lib/utils';

export { useForm, type Control, type SubmitHandler };

export type ControlledInputFieldDescriptor<T extends FieldValues> = {
  id: string;
  name: Path<T>;
  label: string;
  placeholder?: string;
  type?: string;
  inputId?: string;
  visible?: boolean;
};

export function ControlledInputFieldList<T extends FieldValues>({ control, fields }: { control: Control<T>; fields: ControlledInputFieldDescriptor<T>[] }) {
  return <>{fields.filter((field) => field.visible !== false).map((field) => <ControlledInputField key={field.id} control={control} field={field} />)}</>;
}

function ControlledInputField<T extends FieldValues>({ control, field }: { control: Control<T>; field: ControlledInputFieldDescriptor<T> }) {
  return <Controller control={control} name={field.name} render={({ field: input }) => <label className="formField" htmlFor={field.inputId || field.id}><Label>{field.label}</Label><Input {...input} value={String(input.value || '')} id={field.inputId || field.id} type={field.type || 'text'} placeholder={field.placeholder} /></label>} />;
}

export function ControlledTextareaField<T extends FieldValues>({ control, name, className, placeholder }: { control: Control<T>; name: Path<T>; className?: string; placeholder?: string }) {
  return <Controller control={control} name={name} render={({ field }) => <Textarea {...field} value={String(field.value || '')} className={cn('textarea', className)} placeholder={placeholder} />} />;
}
