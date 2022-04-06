import chat from 'k6/x/chat';
import { check } from 'k6';

export default () => {
  const data = { a: 5, b: 2 };
	const addr = "grpcbin.test.k6.io:9000"

  const response = chat.send(addr, "/addsvc.Add/Sum", data)

  check(response, {
    'status is OK': (r) => r && r.status == 0
  });

  console.log(`status: ${JSON.stringify(response.status)}`);
  console.log(`error: ${JSON.stringify(response.error)}`);
  console.log(`message: ${JSON.stringify(response.message)}`);
  console.log(`sum: ${response.message.v}`);
};
