import sys

class Option:
    def __init__(self, aliases, is_flag, default_val):
        self.aliases = aliases
        self.is_flag = is_flag
        self.default_val = default_val
        self.value = []

    def needs_value(self):
        return not self.is_flag

    def set_flag(self):
        if self.is_flag:
            self.value = ["true"]

    def set_value(self, val):
        self.value.append(val)

    def is_set(self):
        return len(self.value) > 0

    def get_value(self):
        if len(self.value) == 0:
            return self.default_val
        return self.value[0]

    def get_values(self):
        if len(self.value) == 0 and self.default_val != "":
            return [self.default_val]
        return self.value

def smart_args(opt_string, args=None):
    if args is None:
        args = sys.argv[1:]

    options = {}
    skipped_args = []

    if not opt_string:
        return options, args

    for opt in opt_string.split(","):
        is_flag = not opt.endswith(":")
        if opt.endswith(":"):
            opt = opt[:-1]
        
        parts = opt.split("=")
        default_val = ""
        if not is_flag and len(parts) > 1:
            default_val = parts[1]
        
        aliases = parts[0].split("|")
        option_obj = Option(aliases, is_flag, default_val)
        
        for alias in aliases:
            options[alias] = option_obj

    i = 0
    while i < len(args):
        arg = args[i]
        
        if arg == "--":
            skipped_args.extend(args[i+1:])
            break
            
        if arg in options:
            if options[arg].needs_value():
                if i + 1 >= len(args):
                    print(f"Missing argument for option {arg}")
                    sys.exit(1)
                i += 1
                options[arg].set_value(args[i])
            else:
                options[arg].set_flag()
        else:
            skipped_args.append(arg)
        
        i += 1
        
    return options, skipped_args
