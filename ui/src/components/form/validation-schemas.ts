import { z } from "zod"

// Project form schema
export const projectFormSchema = z.object({
  name: z.string().min(1, "Project name is required").max(100, "Project name must be less than 100 characters"),
  description: z.string().optional(),
  team: z.string().optional(),
  tags: z.array(z.string()).optional(),
  visibility: z.enum(["public", "private"], {
    required_error: "Please select a visibility option",
  }),
})

// Pipeline form schema
export const pipelineFormSchema = z.object({
  name: z.string().min(1, "Pipeline name is required").max(100, "Pipeline name must be less than 100 characters"),
  description: z.string().optional(),
  projectId: z.string().min(1, "Project is required"),
  schedule: z.string().optional(),
  environment: z.enum(["development", "staging", "production"], {
    required_error: "Please select an environment",
  }),
  tags: z.array(z.string()).optional(),
})

// User profile form schema
export const userProfileFormSchema = z.object({
  firstName: z.string().min(1, "First name is required").max(50, "First name must be less than 50 characters"),
  lastName: z.string().min(1, "Last name is required").max(50, "Last name must be less than 50 characters"),
  email: z.string().email("Please enter a valid email address"),
  bio: z.string().max(500, "Bio must be less than 500 characters").optional(),
  timezone: z.string().optional(),
  emailNotifications: z.boolean().default(true),
  slackNotifications: z.boolean().default(false),
})

// Team form schema
export const teamFormSchema = z.object({
  name: z.string().min(1, "Team name is required").max(100, "Team name must be less than 100 characters"),
  description: z.string().max(500, "Description must be less than 500 characters").optional(),
  members: z.array(z.string()).optional(),
})

// Login form schema
export const loginFormSchema = z.object({
  email: z.string().email("Please enter a valid email address"),
  password: z.string().min(1, "Password is required"),
  rememberMe: z.boolean().default(false),
})

// Register form schema
export const registerFormSchema = z.object({
  firstName: z.string().min(1, "First name is required").max(50, "First name must be less than 50 characters"),
  lastName: z.string().min(1, "Last name is required").max(50, "Last name must be less than 50 characters"),
  email: z.string().email("Please enter a valid email address"),
  password: z.string().min(8, "Password must be at least 8 characters").regex(
    /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/,
    "Password must contain at least one uppercase letter, one lowercase letter, and one number"
  ),
  confirmPassword: z.string(),
  acceptTerms: z.boolean().refine((val) => val === true, "You must accept the terms and conditions"),
}).refine((data) => data.password === data.confirmPassword, {
  message: "Passwords don't match",
  path: ["confirmPassword"],
})

// Settings form schema
export const settingsFormSchema = z.object({
  theme: z.enum(["light", "dark", "system"]),
  language: z.string(),
  timezone: z.string(),
  dateFormat: z.string(),
  timeFormat: z.enum(["12h", "24h"]),
  emailNotifications: z.boolean().default(true),
  desktopNotifications: z.boolean().default(false),
  autoSave: z.boolean().default(true),
  keyboardShortcuts: z.boolean().default(true),
})

export type ProjectFormData = z.infer<typeof projectFormSchema>
export type PipelineFormData = z.infer<typeof pipelineFormSchema>
export type UserProfileFormData = z.infer<typeof userProfileFormSchema>
export type TeamFormData = z.infer<typeof teamFormSchema>
export type LoginFormData = z.infer<typeof loginFormSchema>
export type RegisterFormData = z.infer<typeof registerFormSchema>
export type SettingsFormData = z.infer<typeof settingsFormSchema>