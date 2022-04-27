import { UnleashClient } from 'unleash-proxy-client';

export async function getUnleash() {
    const unleash = new UnleashClient({
        url: 'http://unleash-proxy.testing.com/proxy',
        clientKey: 'clientKeyslkfsdklfkslfd',
        appName: 'default',
        environment: 'development'
    });

// Used to set the context fields, shared with the Unleash Proxy
    await unleash.updateContext({ userId: '1233' });

// Start the background polling
    await unleash.start();

    return unleash
}
