import re

with open('/home/jules/.cache/bazel/_bazel_jules/8c069df52082beee3c95ca17836fb8e2/external/rules_android++android_sdk_repository_extension+androidsdk/helper.bzl', 'r') as f:
    content = f.read()

content = re.sub(r'load\("@local_config_platform//:constraints.bzl", "HOST_CONSTRAINTS"\)', '', content)
content = content.replace('HOST_CONSTRAINTS = []\n', '')

lines = content.split('\n')
loads = [l for l in lines if l.startswith('load(')]
others = [l for l in lines if not l.startswith('load(') and l != '']

new_content = '\n'.join(loads) + '\n\nHOST_CONSTRAINTS = []\n' + '\n'.join(others)

with open('/home/jules/.cache/bazel/_bazel_jules/8c069df52082beee3c95ca17836fb8e2/external/rules_android++android_sdk_repository_extension+androidsdk/helper.bzl', 'w') as f:
    f.write(new_content)
