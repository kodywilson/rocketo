# rocketo
Go utility for interacting with Rocket APEX apis

This is early beta code.

Currently, the app reaches out to retrieve a secret from an OCI vault which is then used
to get a token from an APEX Oauth protected api endpoint. That token is used to auth
calls to Rocket APEX (ORDS) api endpoints.

Future versions will be able to use resource principals, but this version relies on
environment variables. The easiest way to see what to set up is to add an api key
(log into tenancy console, click your user, then user settings, find api key section).
As part of generating an api key, it will print out a basic config. Use this info to
set the required environment variables.

See here for more information.
https://docs.oracle.com/en-us/iaas/Content/API/SDKDocs/cliconfigure.htm

apex url and vault secret ocid are passed in as arguments to the application.
Run the app with no arguments for help text on what to use.
