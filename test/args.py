import argparse

parser = argparse.ArgumentParser(description="runtest")

parser.add_argument('--drawOnly',
                    action='store_true',
                    help='only redraw the test result without run the test')

parser.set_defaults(drawOnly=False)
args = parser.parse_args()

# args result
drawOnly = args.drawOnly
