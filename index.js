exports.handler = async (event) => {
  console.log("Notification Event: ", JSON.stringify(event, null, 2));
  return {
    statusCode: 200,
    body: JSON.stringify({
      message: "Notification Service processed event successfully",
    }),
  };
};
