# gway - application gateway
The gateway acts as a backend-for-frontend. It abstracts the services from the calling frontend. No advanced logic is present, it only calls the backend services by their provided clients (a prerequisite, each service has to provide a client). The technology used is not important. Either http or grpc is valid.

Apart from only calling backend-services the ```gway``` is also responsible for authentication. It implements the OIDC integration with the identity provider and provides the tokens used for all services.
