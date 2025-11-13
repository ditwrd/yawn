import * as React from "react"
import { useForm, UseFormReturn, SubmitHandler } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import { Button } from "@/components/ui/button"
import {
  Form as ShadcnForm,
  FormControl,
  FormDescription,
  FormField as ShadcnFormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form"
import { FormField } from "./form-field"

interface FormFieldConfig {
  name: string
  label?: string
  placeholder?: string
  description?: string
  type?: "text" | "email" | "password" | "textarea" | "select"
  options?: { value: string; label: string }[]
  disabled?: boolean
  className?: string
}

interface ReusableFormProps<T extends z.ZodSchema<any, any>> {
  schema: T
  defaultValues?: z.infer<T>
  fields: FormFieldConfig[]
  onSubmit: SubmitHandler<z.infer<T>>
  submitText?: string
  cancelText?: string
  onCancel?: () => void
  isSubmitting?: boolean
  disabled?: boolean
  className?: string
  children?: (form: UseFormReturn<z.infer<T>>) => React.ReactNode
}

export function ReusableForm<T extends z.ZodSchema<any, any>>({
  schema,
  defaultValues,
  fields,
  onSubmit,
  submitText = "Submit",
  cancelText,
  onCancel,
  isSubmitting = false,
  disabled = false,
  className,
  children,
}: ReusableFormProps<T>) {
  const form = useForm<z.infer<T>>({
    resolver: zodResolver(schema),
    defaultValues,
  })

  return (
    <ShadcnForm {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className={className}>
        <div className="space-y-6">
          {fields.map((fieldConfig) => (
            <FormField
              key={fieldConfig.name}
              control={form.control}
              name={fieldConfig.name as any}
              label={fieldConfig.label}
              placeholder={fieldConfig.placeholder}
              description={fieldConfig.description}
              type={fieldConfig.type}
              options={fieldConfig.options}
              disabled={disabled || fieldConfig.disabled}
              className={fieldConfig.className}
            />
          ))}

          {children?.(form)}

          <div className="flex items-center gap-4">
            <Button
              type="submit"
              disabled={isSubmitting || disabled}
              className="flex-1"
            >
              {isSubmitting ? "Submitting..." : submitText}
            </Button>

            {onCancel && (
              <Button
                type="button"
                variant="outline"
                onClick={onCancel}
                disabled={isSubmitting || disabled}
                className="flex-1"
              >
                {cancelText}
              </Button>
            )}
          </div>
        </div>
      </form>
    </ShadcnForm>
  )
}