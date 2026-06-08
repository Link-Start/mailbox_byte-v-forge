import * as React from "react"
import { ToggleGroup as ToggleGroupPrimitive } from "radix-ui"

import { buttonVariants } from "@/components/ui/button"
import { cn } from "@/lib/utils"

function ToggleGroup({ className, ...props }: React.ComponentProps<typeof ToggleGroupPrimitive.Root>) {
  return <ToggleGroupPrimitive.Root data-slot="toggle-group" className={cn("flex items-center gap-1", className)} {...props} />
}

function ToggleGroupItem({ className, ...props }: React.ComponentProps<typeof ToggleGroupPrimitive.Item>) {
  return (
    <ToggleGroupPrimitive.Item
      data-slot="toggle-group-item"
      className={cn(buttonVariants({ variant: "ghost" }), "h-8 min-w-0 flex-1 px-3 data-[state=on]:bg-background data-[state=on]:shadow-sm", className)}
      {...props}
    />
  )
}

export { ToggleGroup, ToggleGroupItem }
