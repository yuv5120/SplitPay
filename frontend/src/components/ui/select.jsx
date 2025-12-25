import * as React from "react";
import { ChevronDown } from "lucide-react";
import { cn } from "../../lib/utils";

export const Select = ({ value, onValueChange, children }) => {
  const [open, setOpen] = React.useState(false);
  const ref = React.useRef(null);

  React.useEffect(() => {
    const handleClickOutside = (event) => {
      if (ref.current && !ref.current.contains(event.target)) {
        setOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  return (
    <div ref={ref} className="relative">
      {React.Children.map(children, (child) => {
        if (child.type === SelectTrigger) {
          return React.cloneElement(child, { open, setOpen, value });
        }
        if (child.type === SelectContent) {
          return React.cloneElement(child, { open, onValueChange, setOpen });
        }
        return child;
      })}
    </div>
  );
};

export const SelectTrigger = ({ className, children, open, setOpen, value, placeholder }) => {
  // Extract the placeholder from SelectValue child if it exists
  let displayText = placeholder;
  
  React.Children.forEach(children, (child) => {
    if (child && child.type === SelectValue) {
      displayText = child.props.children || child.props.placeholder;
    }
  });

  return (
    <button
      type="button"
      onClick={() => setOpen(!open)}
      className={cn(
        "flex h-10 w-full items-center justify-between rounded-lg border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 transition-all",
        className
      )}
    >
      <span className={!value ? "text-muted-foreground" : ""}>
        {value ? displayText : (placeholder || "Select an option")}
      </span>
      <ChevronDown className="h-4 w-4 opacity-50" />
    </button>
  );
};

export const SelectValue = ({ placeholder, children }) => {
  return children || placeholder;
};

export const SelectContent = ({ className, children, open, onValueChange, setOpen }) => {
  if (!open) return null;

  return (
    <div
      className={cn(
        "absolute z-50 mt-1 max-h-60 w-full overflow-auto rounded-lg border bg-popover text-popover-foreground shadow-lg animate-fade-in",
        className
      )}
    >
      {React.Children.map(children, (child) => {
        if (child.type === SelectItem) {
          return React.cloneElement(child, { onValueChange, setOpen });
        }
        return child;
      })}
    </div>
  );
};

export const SelectItem = ({ className, children, value, onValueChange, setOpen }) => (
  <div
    onClick={() => {
      onValueChange(value);
      setOpen(false);
    }}
    className={cn(
      "relative flex cursor-pointer select-none items-center rounded-sm py-2 px-3 text-sm outline-none hover:bg-accent hover:text-accent-foreground transition-colors",
      className
    )}
  >
    {children}
  </div>
);
